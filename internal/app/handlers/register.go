package handlers

import (
	"net/http"
	"web-forum/internal/app/templates"
)

func HandleRegisterPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.RegisterPage, nil)
}
