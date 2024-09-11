package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func LoginPage(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, _ := Base(stdRequest, stdWriter)
	(*infoToSend)["Title"] = "Авторизация"
	defer templates.Index.Execute(*stdWriter, infoToSend)

	templates.ContentAdd(infoToSend, templates.LoginPage, nil)
}
