package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func FAQPage(reader *http.Request) {
	templates.ContentAdd(reader, templates.FAQ, nil)
}
