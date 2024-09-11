package handlers

import (
	"fmt"
	"log"
	"net/http"
	"web-forum/internal"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

func MainPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	const errorFunction = "handlers.MainPage"

	infoToSend, _ := Base(stdRequest, stdWriter)
	(*infoToSend)["Title"] = internal.SiteName
	defer templates.Index.Execute(*stdWriter, infoToSend)

	categorys, err := category.Get()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, err))
	}

	templates.ContentAdd(infoToSend, templates.Forum, map[string]interface{}{
		"categorys":          categorys,
		"categorys_is_empty": len(*categorys) == 0,
	})
}
