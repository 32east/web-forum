package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func LoginPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.LoginPage, nil)
}
