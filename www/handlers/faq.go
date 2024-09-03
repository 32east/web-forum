package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleFAQPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "FAQ"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.FAQ, nil)
}
