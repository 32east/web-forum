package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleRegisterPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Регистрация"
	defer templates.IndexTemplate.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.RegisterTemplate, nil)
}
