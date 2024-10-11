package admin

import (
	"net/http"
	"strconv"
	"web-forum/api/profile"
	"web-forum/www/services/account"
)

func HandleProfileSettings(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var id = r.FormValue("id")
	var conv, err = strconv.Atoi(id)

	if err != nil {
		answer["success"], answer["reason"] = false, "invalid id"
		return nil
	}

	var accountData, errGetAccount = account.GetById(conv)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "invalid user"
		return nil
	}

	return profile.ProcessSettings(accountData, r, &answer)
}
