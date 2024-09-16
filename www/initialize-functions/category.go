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
			"forum_id":       forumId,
			"forum_name":     output.Name,
			"Description":    output.Description,
			"topics":         topics.Objects,
			"call_paginator": topics.AllPages > 1,
			"current_page":   currentPageInt,
			"paginator":      finalPaginator.PagesArray,
		}

		if finalPaginator.Left.Activated {
			contentToSend["paginator_left"] = finalPaginator.Left.WhichPage
		}

		if finalPaginator.Right.Activated {
			contentToSend["paginator_right"] = finalPaginator.Right.WhichPage
		}

		templates.ContentAdd(r, templates.Topics, contentToSend)
	})
}
