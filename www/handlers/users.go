package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func UsersPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := Base(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Юзеры"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.Users, nil)
}
