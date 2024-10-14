package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"web-forum/internal/app/handlers"
	"web-forum/internal/app/templates"
)

func Push404(writer http.ResponseWriter, reader *http.Request) {
	rCtx := reader.Context()
	http.NotFound(writer, reader)
	rCtx = context.WithValue(rCtx, "BlockExecute", true)
	*reader = *reader.WithContext(rCtx)
}

func Mult(uri string, newFunc func(writer http.ResponseWriter, r *http.Request, id int)) {
	var regExp, err = regexp.Compile(uri)

	if err != nil {
		panic(err)
	}

	endUrl := strings.Split(uri, "/")
	http.HandleFunc("/"+endUrl[1]+"/", func(writer http.ResponseWriter, reader *http.Request) {
		infoToSend, accountData := handlers.Base(reader)

		var ctx = reader.Context()
		ctx = context.WithValue(ctx, "InfoToSend", infoToSend)
		ctx = context.WithValue(ctx, "AccountData", accountData)

		reader = reader.WithContext(ctx)

		var str = regExp.FindString(reader.URL.Path)
		var splitStr = strings.Split(str, "/")

		if len(splitStr) <= 2 {
			http.NotFound(writer, reader)
			return
		}

		var id, convErr = strconv.Atoi(splitStr[2])

		if convErr != nil {
			http.NotFound(writer, reader)
			return
		}

		newFunc(writer, reader, id)

		if reader.Context().Value("BlockExecute") == true {
			return
		}

		templates.Index.Execute(writer, infoToSend)
	})

	// templates.ContentAdd(infoToSend, templates.FAQ, nil)
}
