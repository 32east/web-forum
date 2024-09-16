package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"web-forum/www/handlers"
	"web-forum/www/templates"
)

func Push404(writer http.ResponseWriter, reader *http.Request) {
	rCtx := reader.Context()
	http.NotFound(writer, reader)
	rCtx = context.WithValue(rCtx, "BlockExecute", true)
	*reader = *reader.WithContext(rCtx)
}

func Mult(uri string, newFunc func(writer http.ResponseWriter, r *http.Request, id int)) {
	regExp, err := regexp.Compile(uri)

	if err != nil {
		panic(err)
	}

	endUrl := strings.Split(uri, "/")
	http.HandleFunc("/"+endUrl[1]+"/", func(writer http.ResponseWriter, reader *http.Request) {
		infoToSend, accountData := handlers.Base(reader)

		fmt.Println(reader.Header)
		ctx := reader.Context()
		ctx = context.WithValue(ctx, "InfoToSend", infoToSend)
		ctx = context.WithValue(ctx, "AccountData", accountData)

		reader = reader.WithContext(ctx)

		str := regExp.FindString(reader.URL.Path)
		splitStr := strings.Split(str, "/")

		if len(splitStr) <= 2 {
			http.NotFound(writer, reader)
			return
		}

		id, convErr := strconv.Atoi(splitStr[2])

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
