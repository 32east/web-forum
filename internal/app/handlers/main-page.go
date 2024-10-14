package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/services/category"
	"web-forum/internal/app/templates"
	"web-forum/pkg/stuff"
)

type LastMessage struct {
	TopicId           int
	TopicName         string
	TopicMessageCount int
	TopicCreatedBy    string

	CreatorId  int
	CreateTime string

	MessageCreatorAvatar sql.NullString
	MessageBy            string
}

func MainPage(stdRequest *http.Request) {
	const errorFunction = "handlers.MainPage"

	var categorys, err = category.GetAll()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, err))
	}

	lastMessages, errMessages := db.Postgres.Query(ctx, `
select 
    m.topic_id,
    t.topic_name,
    t.message_count,
    u.username,

    m.account_id,
    m.create_time,
    uMessage.avatar,
    uMessage.username
  from messages as m
  inner join topics as t on m.topic_id = t.id
  inner join users as u on u.id = t.created_by
  inner join users as uMessage on uMessage.id = m.account_id
  where (m.topic_id, m.id) in (
    select topic_id, max(id) 
    from messages 
    where id > (select max(id) - 10000 from messages)
    group by topic_id
  )
  order by m.id desc
  limit 10;
	`)

	var sliceMessages []LastMessage

	if errMessages != nil {
		stuff.ErrLog(errorFunction, errMessages)
	} else {
		for lastMessages.Next() {
			var createTime time.Time
			var lastMessage = LastMessage{}

			err = lastMessages.Scan(&lastMessage.TopicId, &lastMessage.TopicName, &lastMessage.TopicMessageCount, &lastMessage.TopicCreatedBy, &lastMessage.CreatorId, &createTime, &lastMessage.MessageCreatorAvatar, &lastMessage.MessageBy)

			if err != nil {
				stuff.ErrLog(errorFunction, errMessages)
				continue
			}

			lastMessage.TopicMessageCount -= 1
			lastMessage.CreateTime = createTime.Format("2006-01-02 15:04:05")
			sliceMessages = append(sliceMessages, lastMessage)
		}
	}

	templates.ContentAdd(stdRequest, templates.Forum, map[string]interface{}{
		"Categorys":    categorys,
		"LastMessages": sliceMessages,
	})
}
