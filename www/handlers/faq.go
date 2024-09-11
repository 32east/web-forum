package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func FAQPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := Base(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "FAQ"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.FAQ, nil)
}
