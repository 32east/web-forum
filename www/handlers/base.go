package handlers

import (
	"context"
	"net/http"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func Base(stdRequest *http.Request) (map[string]interface{}, *account.Account) {
	// go account.TokensRefreshInRedis(stdRequest, writer) // TODO: Расскомент!!!

	infoToSend := make(map[string]interface{})
	cookie, err := stdRequest.Cookie("access_token")

	infoToSend["Authorized"] = false
	accountData := &account.Account{}

	if err != nil {
		return infoToSend, accountData
	}

	accountData, errGetAccount := account.ReadFromCookie(cookie)

	if errGetAccount != nil {
		return infoToSend, accountData
	}

	infoToSend["Authorized"] = true
	infoToSend["AccountId"] = accountData.Id
	infoToSend["Username"] = accountData.Username

	if accountData.Avatar.Valid {
		infoToSend["Avatar"] = accountData.Avatar.String
	}

	return infoToSend, accountData
}
