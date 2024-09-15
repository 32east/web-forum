package handlers

import (
	"fmt"
	"net/http"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

func TopicCreate(stdRequest *http.Request) {
	forums, err := category.Get()

	if err != nil {
		panic(err)
	}

	var categorys []interface{}
	currentCategory := stdRequest.FormValue("category")

	for _, output := range *forums {
		forumId := output.Id

		categorys = append(categorys, map[string]interface{}{
			"forum_name":  output.Name,
			"forum_id":    output.Id,
			"is_selected": fmt.Sprint(forumId) == currentCategory,
		})
	}

	templates.ContentAdd(stdRequest, templates.CreateNewTopic, map[string]interface{}{
		"categorys": categorys,
	})
}
