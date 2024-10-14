package middleware

import (
	"context"
	"net/http"
	"web-forum/internal/app/handlers"
	"web-forum/internal/app/templates"
)

func AdminPage(uri string, title string, newFunc func(*http.Request)) {
	http.HandleFunc(uri, func(writer http.ResponseWriter, reader *http.Request) {
		if uri == "/" && reader.URL.Path != "/" {
			http.NotFound(writer, reader)
			return
		}

		var infoToSend, accountData = handlers.Base(reader)

		if !(*accountData).IsAdmin {
			http.NotFound(writer, reader)
			return
		}

		infoToSend["Title"] = title

		var ctx = reader.Context()
		ctx = context.WithValue(ctx, "InfoToSend", infoToSend)
		ctx = context.WithValue(ctx, "AccountData", *accountData)
		ctx = context.WithValue(ctx, "Writer", writer)

		reader = reader.WithContext(ctx)

		newFunc(reader)

		if reader.Context().Value("BlockExecute") == true {
			return
		}

		templates.Index.Execute(writer, infoToSend)
	})

	// templates.ContentAdd(infoToSend, templates.FAQ, nil)
}
