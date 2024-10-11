package initialize_functions

import (
	"net/http"
	"strconv"
	"web-forum/www/middleware"
	"web-forum/www/services/category"
	"web-forum/www/services/topics"
	"web-forum/www/templates"
)

func Categorys() {
	middleware.Mult("/category/([0-9]+)", func(w http.ResponseWriter, r *http.Request, forumId int) {
		var output = category.GetInfo(forumId)
		var currentPage = r.FormValue("page")

		if currentPage == "" {
			currentPage = "1"
		}

		var currentPageInt, errInt = strconv.Atoi(currentPage)

		if errInt != nil {
			currentPageInt = 0
		}
		var finalPaginator = topics.Get(forumId, currentPageInt)

		contentToSend := map[string]interface{}{
			"Id":                   forumId,
			"Name":                 output.Name,
			"Description":          output.Description,
			"Topics":               finalPaginator.Objects,
			"PaginatorIsActivated": finalPaginator.AllPages > 1,
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
