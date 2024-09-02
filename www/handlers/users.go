package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleUsersPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Юзеры"
	defer templates.IndexTemplate.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.UsersTemplate, nil)
}
