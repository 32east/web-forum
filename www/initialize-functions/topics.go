package initialize_functions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/middleware"
	"web-forum/www/services/account"
	"web-forum/www/services/paginator"
	topics_messages "web-forum/www/services/topics-messages"
	"web-forum/www/templates"
)

func Topics() {
	const errorFunc = "InitializeTopicsPages"
	rows, err := db.Postgres.Query(context.Background(), "SELECT * FROM topics;")
	defer rows.Close()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunc, err))
	}

	for rows.Next() {
		topic := internal.Topic{}
		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount)

		if scanErr != nil {
			log.Fatal(fmt.Errorf("%s [2]: %w", errorFunc, scanErr))
		}

		CreateTopic(topic)
	}
}

// TODO: Лучше это запихнуть в api/topics.
func CreateTopic(inputTopic internal.Topic) string {
	url := "/topics/" + fmt.Sprint(inputTopic.Id) + "/"

	middleware.Page(url, inputTopic.Name, func(r *http.Request) {
		currentPage := r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		page, errInt := strconv.Atoi(currentPage)

		if errInt != nil {
			page = 1
		}

		paginatorMessages, _ := topics_messages.Get(inputTopic, page)
		finalPaginator := paginator.Construct(*paginatorMessages)
		getAccount, ok := account.GetById(inputTopic.Creator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении информации о создателе топика", inputTopic.Creator)
		}

		topicInfo := map[string]interface{}{
			"topic_id":       inputTopic.Id,
			"topic_name":     inputTopic.Name,
			"forum_name":     inputTopic.ForumId,
			"username":       getAccount.Username,
			"create_time":    inputTopic.CreateTime.Format("2006-01-02 15:04:05"),
			"messages":       paginatorMessages.Objects,
			"call_paginator": paginatorMessages.AllPages > 1,
			"current_page":   page,
			"paginator":      finalPaginator.PagesArray,
		}

		if finalPaginator.Left.Activated {
			topicInfo["paginator_left"] = finalPaginator.Left.WhichPage
		}

		if finalPaginator.Right.Activated {
			topicInfo["paginator_right"] = finalPaginator.Right.WhichPage
		}

		if getAccount.Avatar.Valid {
			topicInfo["avatar"] = getAccount.Avatar.String
		}

		if getAccount.SignText.Valid {
			topicInfo["sign_text"] = getAccount.SignText.String
		}

		templates.ContentAdd(r, templates.TopicPage, topicInfo)
	})

	return url
}
