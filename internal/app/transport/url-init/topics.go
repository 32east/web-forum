package url_init

import (
	"fmt"
	"net/http"
	"strconv"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/middleware"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
	topics_messages "web-forum/internal/app/services/topics-messages"
	"web-forum/internal/app/templates"
	"web-forum/pkg/stuff"
)

func Topics() {
	const errorFunc = "Topics"

	middleware.Mult("/topics/([0-9]+)", func(w http.ResponseWriter, r *http.Request, topicId int) {
		var topic = models.Topic{}
		var scanErr = db.Postgres.QueryRow(ctx, "select * from topics where id = $1;", topicId).Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount, &topic.ParentId)

		if scanErr != nil {
			middleware.Push404(w, r)
			return
		}

		topic.MessageCount -= 1
		var currentPage = r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		var page, errInt = strconv.Atoi(currentPage)

		if errInt != nil {
			page = 1
		}

		var finalPaginator = topics_messages.Get(&topic, page)
		var getAccount, ok = account.GetById(topic.Creator)

		if ok != nil {
			stuff.ErrLog(errorFunc, fmt.Errorf("cannot find account: %d", topic.Creator))
		}

		topicInfo := map[string]interface{}{
			"Id":                   topic.Id,
			"Name":                 topic.Name,
			"ForumName":            topic.ForumId,
			"CreatorUsername":      getAccount.Username,
			"CreateTime":           topic.CreateTime.Format("2006-01-02 15:04:05"),
			"Messages":             finalPaginator.Objects,
			"PaginatorIsActivated": finalPaginator.AllPages > 1,
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
