package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"web-forum/pkg/stuff"
)

func API(uri string, newFunc func(http.ResponseWriter, *http.Request, map[string]interface{}) error) {
	http.HandleFunc(uri, func(writer http.ResponseWriter, reader *http.Request) {
		log.Println("Request:", reader.Method, reader.URL.Path)

		writer.Header().Add("content-type", "application/json")

		var errFunction = fmt.Sprintf("%s %s", reader.Method, reader.URL.Path)
		var newJSONEncoder = json.NewEncoder(writer)
		var answer = make(map[string]interface{})
		defer newJSONEncoder.Encode(answer)

		if reader.Method != "POST" {
			answer["success"], answer["reason"] = false, "method not allowed"
			return
		}

		err := newFunc(writer, reader, answer)

		if err != nil {
			stuff.ErrLog(errFunction, err)
		}
	})
}
