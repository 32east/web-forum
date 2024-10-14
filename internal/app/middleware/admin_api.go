package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"web-forum/internal/app/services/account"
	"web-forum/pkg/stuff"
)

func AdminAPI(uri string, method string, newFunc func(http.ResponseWriter, *http.Request, map[string]interface{}) error) {
	http.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)

		w.Header().Add("content-type", "application/json")

		var errFunction = fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		var newJSONEncoder = json.NewEncoder(w)
		var answer = make(map[string]interface{})
		defer newJSONEncoder.Encode(answer)

		if r.Method != method {
			answer["success"], answer["reason"] = false, "method not allowed"
			return
		}

		var cookie, err = r.Cookie("access_token")

		if err != nil {
			answer["success"], answer["reason"] = false, "not authorized"
			return
		}

		var accountData, errGetAccount = account.ReadFromCookie(cookie)

		if errGetAccount != nil {
			answer["success"], answer["reason"] = false, "not authorized"
			return
		}

		if !accountData.IsAdmin {
			answer["success"], answer["reason"] = false, "no access"
			return
		}

		errFunc := newFunc(w, r, answer)

		if errFunc != nil {
			stuff.ErrLog(errFunction, errFunc)
		}
	})
}
