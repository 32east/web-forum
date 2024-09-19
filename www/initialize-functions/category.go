package initialize_functions

import (
	"net/http"
	"strconv"
	"web-forum/www/middleware"
	"web-forum/www/services/category"
	"web-forum/www/services/paginator"
	"web-forum/www/services/topics"
	"web-forum/www/templates"
)

func Categorys() {
	middleware.Mult("/category/([0-9]+)", func(w http.ResponseWriter, r *http.Request, forumId int) {
		output := category.GetInfo(forumId)

		currentPage := r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		currentPageInt, errInt := strconv.Atoi(currentPage)

		if errInt != nil {
			currentPageInt = 0
		}
		topics, _ := topics.Get(forumId, currentPageInt)
		finalPaginator := paginator.Construct(*topics)

		contentToSend := map[string]interface{}{
			"Id":                   forumId,
			"Name":                 output.Name,
			"Description":          output.Description,
			"Topics":               topics.Objects,
			"PaginatorIsActivated": topics.AllPages > 1,
			"Paginator":            finalPaginator.PagesArray,
			"CurrentPage":          currentPageInt,
		}

		if finalPaginator.Left.Activated {
			contentToSend["PaginatorLeft"] = finalPaginator.Left.WhichPage
		}

		if finalPaginator.Right.Activated {
			contentToSend["PaginatorRight"] = finalPaginator.Right.WhichPage
		}

		templates.ContentAdd(r, templates.Topics, contentToSend)
	})
}
