package handlers

import (
	"log"
	"net/http"
	"web-forum/internal"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

func HandleMainPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = internal.SiteName
	defer templates.Index.Execute(*stdWriter, infoToSend)

	categorys, err := category.GetForums()

	if err != nil {
		log.Fatal(err)
	}

	templates.ContentAdd(infoToSend, templates.Forum, map[string]interface{}{
		"categorys":          categorys,
		"categorys_is_empty": len(*categorys) == 0,
	})
}
