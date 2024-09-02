package handlers

import (
	"fmt"
	"net/http"
	"web-forum/system/sqlDb"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

func HandleTopicCreate(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Создание нового топика"
	defer templates.IndexTemplate.Execute(*stdWriter, infoToSend)

	forums, err := category.GetForums(sqlDb.MySqlDB)

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

	templates.ContentAdd(infoToSend, templates.CreateNewTopicTemplate, map[string]interface{}{
		"categorys": categorys,
	})
}
