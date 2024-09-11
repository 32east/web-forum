package topics_messages

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func Get(topic internal.Topic, page int) (*internal.Paginator, error) {
	const errorFunction = "topics_messages.Get"
	tx, err := db.Postgres.Begin(ctx)

	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	var topicsCount float64
	queryRow := tx.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE topic_id = $1;", topic.Id)
	countMessagesErr := queryRow.Scan(&topicsCount)
	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if countMessagesErr != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, countMessagesErr))
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM messages where topic_id=$1 LIMIT %d OFFSET %d;", internal.MaxPaginatorMessages, (page-1)*internal.MaxPaginatorMessages)
	rows, err := tx.Query(ctx, fmtQuery, topic.Id)
	defer rows.Close()

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

		getAccount, ok := account.GetById(accountId)

		if ok != nil {
			// TODO: Создавать нормально транзакцию?
			db.Postgres.Exec(ctx, "DELETE FROM messages WHERE id = $1;", id)

			continue
		}

		if updateTime != nil && updateTime.(sql.NullTime).Valid {
			updateTime = updateTime.(time.Time).Format("2006-01-02 15:04:05")
		}

		messageInfo := map[string]interface{}{
			"uid":         getAccount.Id,
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
