package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func UsersPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.Users, nil)
}
