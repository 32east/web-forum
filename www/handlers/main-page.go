package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

type LastMessage struct {
	TopicId           int
	TopicName         string
	TopicMessageCount int
	TopicCreatedBy    string

	CreatorId  int
	CreateTime string
}

func MainPage(stdRequest *http.Request) {
	const errorFunction = "handlers.MainPage"

	categorys, err := category.GetAll()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, err))
	}

	startTime := time.Now()
	fmt.Print("Querying main page... ")
	lastMessages, errMessages := db.Postgres.Query(ctx, `
select 
    m.topic_id,
    t.topic_name,
    t.message_count,
    u.username,

    m.account_id,
    m.create_time
  from messages as m
  inner join topics as t on m.topic_id = t.id
  inner join users as u on u.id = t.created_by
  where (m.topic_id, m.id) in (
    select topic_id, max(id) 
    from messages 
    where id > (select max(id) - 10000 from messages)
    group by topic_id
  )
  order by m.id desc
  limit 10;
	`)
	fmt.Printf("> %dms\n", time.Now().Sub(startTime).Milliseconds())
	var sliceMessages []LastMessage

	if errMessages != nil {
		system.ErrLog(errorFunction, errMessages)
	} else {
		for lastMessages.Next() {
			var createTime time.Time
			lastMessage := LastMessage{}

			err = lastMessages.Scan(&lastMessage.TopicId, &lastMessage.TopicName, &lastMessage.TopicMessageCount, &lastMessage.TopicCreatedBy, &lastMessage.CreatorId, &createTime)

			if err != nil {
				system.ErrLog(errorFunction, errMessages)
				continue
			}

			lastMessage.TopicMessageCount -= 1
			lastMessage.CreateTime = createTime.Format("2006-01-02 15:04:05")
			sliceMessages = append(sliceMessages, lastMessage)
		}
	}

	templates.ContentAdd(stdRequest, templates.Forum, map[string]interface{}{
		"categorys":          categorys,
		"categorys_is_empty": len(*categorys) == 0,
		"LastMessages":       sliceMessages,
	})
}
