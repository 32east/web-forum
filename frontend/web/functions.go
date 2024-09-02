package web

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
	"web-forum/internal"
	"web-forum/www/services/account"
)

func GetTopics(forumId int, db *sql.DB, page int) (*[]interface{}, int, error) {
	var topics []interface{}

	tx, err := db.Begin()
	defer tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	var topicsCount float64
	queryRow := tx.QueryRow("SELECT COUNT(*) FROM `topics` WHERE forum_id=?;", forumId)
	countTopicsErr := queryRow.Scan(&topicsCount)
	topicsCount = math.Ceil(topicsCount / internal.MaxPaginatorTopics)

	if countTopicsErr != nil {
		log.Fatal(countTopicsErr)
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM `topics` WHERE forum_id = ? ORDER BY id LIMIT %d OFFSET %d;", internal.MaxPaginatorTopics, (page-1)*internal.MaxPaginatorTopics)
	rows, err := tx.Query(fmtQuery, forumId)

	if err != nil {
		log.Fatal(err)
		return &topics, 0, fmt.Errorf("error getting topics: %v", err)
	}

	for rows.Next() {
		var id int
		var forumId int
		var topicName string
		var topicMessage string
		var topicCreator int
		var createTime time.Time
		var updateTime interface{}

		scanErr := rows.Scan(&id, &forumId, &topicName, &topicMessage, &topicCreator, &createTime, &updateTime)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		if updateTime != nil && updateTime.(sql.NullTime).Valid {
			updateTime = updateTime.(time.Time).Format("2006-01-02 15:04:05")
		}

		creatorAccount, ok := account.GetAccountById(topicCreator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении креатор аккаунта", topicCreator)
		}

		aboutTopic := map[string]interface{}{
			"topic_id":    id,
			"forum_id":    forumId,
			"topic_name":  topicName,
			"username":    creatorAccount.Username,
			"create_time": createTime.Format("2006-01-02 15:04:05"),
			"update_time": updateTime,
		}

		if creatorAccount.Avatar.Valid {
			aboutTopic["avatar"] = creatorAccount.Avatar.String
		}

		topics = append(topics, aboutTopic)
	}

	return &topics, int(topicsCount), nil
}

func Test_AddTopic(db *sql.DB, forumId int, topicName string, topicMessage string, topicCreator int) {
	_, err := db.Query("INSERT INTO `topics` (id, forum_id, topic_name, topic_message, created_by, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, ?, NULL)", forumId, topicName, topicMessage, topicCreator, time.Now())

	if err != nil {
		log.Fatal(err)
	}
}
