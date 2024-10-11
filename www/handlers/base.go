package handlers

import (
	"context"
	"net/http"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func Base(stdRequest *http.Request) (map[string]interface{}, *account.Account) {
	var infoToSend = make(map[string]interface{})
	var cookie, err = stdRequest.Cookie("access_token")

	infoToSend["Authorized"] = false
	var accountData = &account.Account{}

	if err != nil {
		return infoToSend, accountData
	}

	accountData, errGetAccount := account.ReadFromCookie(cookie)

	if errGetAccount != nil {
		return infoToSend, accountData
	}

	if accountData.IsAdmin {
		infoToSend["IsAdmin"] = true
	}

	infoToSend["Authorized"] = true
	infoToSend["AccountId"] = accountData.Id
	infoToSend["Username"] = accountData.Username

	if accountData.Avatar.Valid {
		infoToSend["Avatar"] = accountData.Avatar.String
	}

	return infoToSend, accountData
}
