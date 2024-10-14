package topics

import (
	"context"
	"fmt"
	"strconv"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
	"web-forum/internal/app/services/paginator"
	"web-forum/pkg/stuff"
)

var ctx = context.Background()

func Get(forumId int, page int) *models.Paginator {
	const errorFunction = "topics.Get"

	var preparedValue int
	var sqlCount string
	var topics = models.Paginator{}

	if forumId == -1 {
		usersCount, usersCountErr := rdb.RedisDB.Get(ctx, "count:topics").Result()

		if usersCountErr != nil {
			stuff.ErrLog(errorFunction, usersCountErr)

			usersCount = "0"
		}

		conv, err := strconv.Atoi(usersCount)

		if err != nil {
			stuff.ErrLog(errorFunction, err)
			conv = 1
		}

		preparedValue = conv
	} else {
		sqlCount = fmt.Sprintf("select topics_count from categorys where id = %d;", forumId)
	}

	preQuery := models.PaginatorPreQuery{
		TableName:     "topics",
		OutputColumns: "id, forum_id, topic_name, created_by, create_time, update_time, message_count, parent_id",
		WhereColumn:   "forum_id",
		WhereValue:    forumId,
		Page:          page,
		OrderReverse:  true,
		QueryCount: models.PaginatorQueryCount{
			PreparedValue: preparedValue,
			Query:         sqlCount,
		},
	}

	tx, rows, topics, err := paginator.Query(preQuery)
	if tx != nil {
		defer tx.Commit(ctx)
	}

	if err != nil {
		return &topics
	}

	var tempUsers []int
	var tempTopics []models.Topic

	for rows.Next() {
		topic := models.Topic{}

		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount, &topic.ParentId)

		if scanErr != nil {
			stuff.ErrLog(errorFunction, scanErr)
			continue
		}

		topic.MessageCount -= 1

		tempUsers = append(tempUsers, topic.Creator)
		tempTopics = append(tempTopics, topic)
	}

	var usersInfo = account.GetFromSlice(tempUsers, tx)

	for i := 0; i < len(tempTopics); i++ {
		var topic = tempTopics[i]
		var updateTime = ""

		if topic.UpdateTime.Valid {
			updateTime = topic.UpdateTime.Time.Format("2006-01-02 15:04:05")
		}

		var creatorAccount, ok = usersInfo[topic.Creator]

		if !ok {
			stuff.ErrLog("topics.Get", fmt.Errorf("Не найден креатор топика в БД?"))
			continue
		}

		aboutTopic := map[string]interface{}{
			"Id":              topic.Id,
			"Name":            topic.Name,
			"ForumId":         forumId,
			"CreatorUsername": creatorAccount.Username,
			"CreateTime":      topic.CreateTime.Format("2006-01-02 15:04:05"),
			"UpdateTime":      updateTime,

			// Поскольку 1 сообщение - это сообщение самого топика.
			"MessageCount": topic.MessageCount,
		}

		if creatorAccount.Avatar.Valid {
			aboutTopic["Avatar"] = creatorAccount.Avatar.String
		}

		topics.Objects = append(topics.Objects, aboutTopic)
	}

	return &topics
}
