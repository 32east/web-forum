package topics

import (
	"fmt"
	"log"
	"math"
	"web-forum/internal"
	"web-forum/system/sqlDb"
	"web-forum/www/services/account"
)

func GetTopics(forumId int, page int) (*internal.Paginator, error) {
	var topics internal.Paginator

	tx, err := sqlDb.MySqlDB.Begin()
	defer tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	var topicsCount float64
	queryRow := tx.QueryRow("SELECT COUNT(*) FROM `topics` WHERE forum_id=?;", forumId)
	countTopicsErr := queryRow.Scan(&topicsCount)
	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorTopics)

	if countTopicsErr != nil {
		log.Fatal(countTopicsErr)
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM `topics` WHERE forum_id = ? ORDER BY id LIMIT %d OFFSET %d;", internal.MaxPaginatorTopics, (page-1)*internal.MaxPaginatorTopics)
	rows, err := tx.Query(fmtQuery, forumId)

	if err != nil {
		log.Fatal("[functions:35]", err)
	}

	for rows.Next() {
		topic := internal.Topic{}

		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		updateTime := ""

		if topic.UpdateTime.Valid {
			updateTime = topic.UpdateTime.Time.Format("2006-01-02 15:04:05")
		}

		creatorAccount, ok := account.GetAccountById(topic.Creator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении креатор аккаунта", topic.Creator)
		}

		aboutTopic := map[string]interface{}{
			"topic_id":    topic.Id,
			"forum_id":    forumId,
			"topic_name":  topic.Name,
			"username":    creatorAccount.Username,
			"create_time": topic.CreateTime.Format("2006-01-02 15:04:05"),
			"update_time": updateTime,

			// Поскольку 1 сообщение - это сообщение самого топика.
			"message_count": topic.MessageCount,
		}

		if creatorAccount.Avatar.Valid {
			aboutTopic["avatar"] = creatorAccount.Avatar.String
		}

		topics.Objects = append(topics.Objects, aboutTopic)
	}

	topics.CurrentPage = page
	topics.AllPages = int(pagesCount)

	return &topics, nil
}
