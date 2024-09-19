package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"web-forum/system"
	"web-forum/www/services/account"
)

func AdminAPI(uri string, method string, newFunc func(http.ResponseWriter, *http.Request, map[string]interface{})) {
	http.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)

		var errFunction = fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		w.Header().Add("content-type", "application/json")

		newJSONEncoder := json.NewEncoder(w)
		answer := make(map[string]interface{})
		defer newJSONEncoder.Encode(answer)

		if r.Method != method {
			answer["success"], answer["reason"] = false, "method not allowed"
			return
		}

		cookie, err := r.Cookie("access_token")

		if err != nil {
			answer["success"], answer["reason"] = false, "not authorized"
			return
		}

		accountData, errGetAccount := account.ReadFromCookie(cookie)

		if errGetAccount != nil {
			answer["success"], answer["reason"] = false, "not authorized"
			return
		}

		if !accountData.IsAdmin {
			answer["success"], answer["reason"] = false, "no access"
			return
		}

		newFunc(w, r, answer)

		if answer["success"] != nil && !answer["success"].(bool) {
			system.ErrLog(errFunction, fmt.Errorf("%s", answer["reason"]))
		}
	})
}
