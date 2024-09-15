package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"web-forum/system"
)

func API(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, reader *http.Request) {
		log.Println("Request:", reader.Method, reader.URL.Path)

		const errFunction = "HandleRegister"

		header := writer.Header()
		header.Add("content-type", "application/json")

		newJSONEncoder := json.NewEncoder(writer)
		answer := make(map[string]interface{})

		defer newJSONEncoder.Encode(answer)
		defer func() {
			if !answer["success"].(bool) {
				system.ErrLog(errFunction, fmt.Errorf(string(reader.RemoteAddr)+" > "+answer["reason"].(string)))
			}
		}()

		if reader.Method != "POST" {
			answer["success"], answer["reason"] = false, "method not allowed"
			return
		}

		next.ServeHTTP(writer, reader)
	})
}
