package topics_messages

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
	"web-forum/internal"
	"web-forum/system/sqlDb"
	"web-forum/www/services/account"
)

func GetMessages(topic internal.Topic, page int) (*internal.Paginator, error) {
	tx, err := sqlDb.MySqlDB.Begin()
	defer tx.Commit()

	var topicsCount float64
	queryRow := tx.QueryRow("SELECT COUNT(*) FROM `messages` WHERE topic_id=?;", topic.Id)
	countMessagesErr := queryRow.Scan(&topicsCount)
	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if countMessagesErr != nil {
		log.Fatal(countMessagesErr)
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM `messages` where topic_id=? LIMIT %d OFFSET %d;", internal.MaxPaginatorMessages, (page-1)*internal.MaxPaginatorMessages)
	rows, err := tx.Query(fmtQuery, topic.Id)

	if err != nil {
		return nil, err
	}

	paginatorMessages := internal.Paginator{
		CurrentPage: page,
		AllPages:    int(pagesCount),
	}

	for rows.Next() {
		var id int
		var topicId int
		var accountId int
		var message string
		var createTime time.Time
		var updateTime interface{}

		rows.Scan(&id, &topicId, &accountId, &message, &createTime, &updateTime)

		getAccount, ok := account.GetAccountById(accountId)

		if ok != nil {
			// TODO: Создавать нормально транзакцию?
			sqlDb.MySqlDB.Query("DELETE FROM `messages` WHERE `id`=?;", id)

			continue
		}

		if updateTime != nil && updateTime.(sql.NullTime).Valid {
			updateTime = updateTime.(time.Time).Format("2006-01-02 15:04:05")
		}

		messageInfo := map[string]interface{}{
			"username":    getAccount.Username,
			"message":     message,
			"create_time": createTime.Format("2006-01-02 15:04:05"),
			"update_time": updateTime,
		}

		if getAccount.Avatar.Valid {
			messageInfo["avatar"] = getAccount.Avatar.String
		}

		if getAccount.SignText.Valid {
			messageInfo["sign_text"] = getAccount.SignText.String
		}

		paginatorMessages.Objects = append(paginatorMessages.Objects, messageInfo)
	}

	return &paginatorMessages, nil
}
