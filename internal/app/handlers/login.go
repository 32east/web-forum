package handlers

import (
	"net/http"
	"web-forum/internal/app/templates"
)

func LoginPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.LoginPage, nil)
}
