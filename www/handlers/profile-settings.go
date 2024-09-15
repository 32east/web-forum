package handlers

import (
	"net/http"
	account2 "web-forum/www/services/account"
	"web-forum/www/templates"
)

func HandleProfileSettings(stdRequest *http.Request) {
	reqCtx := stdRequest.Context()
	account := reqCtx.Value("AccountData").(*account2.Account)
	infoToSend := reqCtx.Value("InfoToSend").(map[string]interface{})
	authorized := infoToSend["Authorized"]

	if !authorized.(bool) {
		infoToSend["Title"] = "Нет доступа"
		templates.ContentAdd(stdRequest, templates.ProfileSettings, nil)
		return
	}

	infoToSend["Title"] = "Настройки профиля"

	contentToAdd := map[string]interface{}{}

	contentToAdd["Email"] = account.Email
	contentToAdd["Username"] = account.Username

	if account.Description.Valid {
		contentToAdd["Description"] = account.Description.String
	}

	if account.SignText.Valid {
		contentToAdd["SignText"] = account.SignText.String
	}

	if account.Avatar.Valid {
		contentToAdd["Avatar"] = account.Avatar.String
	}

	templates.ContentAdd(stdRequest, templates.ProfileSettings, contentToAdd)
}
