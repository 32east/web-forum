package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleRegisterPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.RegisterPage, nil)
}
