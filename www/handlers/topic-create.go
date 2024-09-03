package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web-forum/internal"
	"web-forum/www/services/account"
	"web-forum/www/services/category"
	"web-forum/www/services/paginator"
	topics_messages "web-forum/www/services/topics-messages"
	"web-forum/www/templates"
)

func HandleTopicCreate(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Создание нового топика"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	forums, err := category.GetForums()

	if err != nil {
		panic(err)
	}

	var categorys []interface{}
	currentCategory := stdRequest.FormValue("category")

	for _, output := range *forums {
		outputToMap := output.(map[string]interface{})
		forumId := outputToMap["forum_id"]

		categorys = append(categorys, map[string]interface{}{
			"forum_name":  outputToMap["forum_name"].(string),
			"forum_id":    outputToMap["forum_id"].(int),
			"is_selected": fmt.Sprint(forumId) == currentCategory,
		})
	}

	templates.ContentAdd(infoToSend, templates.CreateNewTopic, map[string]interface{}{
		"categorys": categorys,
	})
}

func CreateTopic(inputTopic internal.Topic) string {
	url := "/topics/" + fmt.Sprint(inputTopic.Id) + "/"

	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		infoToSend, _ := HandleBase(r, &w)
		(*infoToSend)["Title"] = inputTopic.Name
		defer templates.Index.Execute(w, infoToSend)

		currentPage := r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		page, errInt := strconv.Atoi(currentPage)

		if errInt != nil {
			page = 0
		}

		paginatorMessages, _ := topics_messages.GetMessages(inputTopic, page)
		finalPaginator := paginator.PaginatorConstruct(*paginatorMessages)
		getAccount, ok := account.GetAccountById(inputTopic.Creator)

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

		templates.ContentAdd(infoToSend, templates.TopicPage, topicInfo)
	})

	return url
}
