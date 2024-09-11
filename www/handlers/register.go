package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleRegisterPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := Base(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Регистрация"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.RegisterPage, nil)
}
