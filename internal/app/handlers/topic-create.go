package handlers

import (
	"fmt"
	"net/http"
	"web-forum/internal/app/services/category"
	"web-forum/internal/app/templates"
)

func TopicCreate(stdRequest *http.Request) {
	var categoryList, err = category.GetAll()

	if err != nil {
		panic(err)
	}

	var categorys []interface{}
	var currentCategory = stdRequest.FormValue("category")

	for _, output := range *categoryList {
		var forumId = output.Id

		categorys = append(categorys, map[string]interface{}{
			"Id":         output.Id,
			"Name":       output.Name,
			"IsSelected": fmt.Sprint(forumId) == currentCategory,
		})
	}

	templates.ContentAdd(stdRequest, templates.CreateNewTopic, map[string]interface{}{
		"Categorys": categorys,
	})
}
