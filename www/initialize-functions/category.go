package initialize_functions

import (
	"fmt"
	"net/http"
	"strconv"
	"web-forum/www/handlers"
	"web-forum/www/services/category"
	"web-forum/www/services/paginator"
	"web-forum/www/services/topics"
	"web-forum/www/templates"
)

func Categorys() {
	forums, err := category.Get()

	if err != nil {
		panic(err)
	}

	for _, output := range *forums {
		forumId := output.Id

		http.HandleFunc("/category/"+fmt.Sprint(forumId)+"/", func(w http.ResponseWriter, r *http.Request) {
			currentPage := r.FormValue("page")

			infoToSend, _ := handlers.Base(r, &w)
			(*infoToSend)["Title"] = output.Name
			defer templates.Index.Execute(w, infoToSend)

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

			templates.ContentAdd(infoToSend, templates.Topics, contentToSend)
		})
	}
}
