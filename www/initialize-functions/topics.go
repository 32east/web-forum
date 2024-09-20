package initialize_functions

import (
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

	middleware.Mult("/topics/([0-9]+)", func(w http.ResponseWriter, r *http.Request, topicId int) {
		topic := internal.Topic{}
		scanErr := db.Postgres.QueryRow(ctx, "select * from topics where id = $1;", topicId).Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount, &topic.ParentId)

		if scanErr != nil {
			middleware.Push404(w, r)
			return
		}

		topic.MessageCount -= 1
		currentPage := r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		page, errInt := strconv.Atoi(currentPage)

		if errInt != nil {
			page = 1
		}

		paginatorMessages, _ := topics_messages.Get(topic, page)
		finalPaginator := paginator.Construct(*paginatorMessages)
		getAccount, ok := account.GetById(topic.Creator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении информации о создателе топика", topic.Creator)
		}

		topicInfo := map[string]interface{}{
			"Id":                   topic.Id,
			"Name":                 topic.Name,
			"ForumName":            topic.ForumId,
			"CreatorUsername":      getAccount.Username,
			"CreateTime":           topic.CreateTime.Format("2006-01-02 15:04:05"),
			"Messages":             paginatorMessages.Objects,
			"PaginatorIsActivated": paginatorMessages.AllPages > 1,
			"Paginator":            finalPaginator.PagesArray,
			"CurrentPage":          page,
		}

		if finalPaginator.Left.Activated {
			topicInfo["PaginatorLeft"] = finalPaginator.Left.WhichPage
		}

		if finalPaginator.Right.Activated {
			topicInfo["PaginatorRight"] = finalPaginator.Right.WhichPage
		}

		if getAccount.Avatar.Valid {
			topicInfo["Avatar"] = getAccount.Avatar.String
		}

		if getAccount.SignText.Valid {
			topicInfo["SignText"] = getAccount.SignText.String
		}

		templates.ContentAdd(r, templates.TopicPage, topicInfo)
	})
}
