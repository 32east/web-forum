package handlers

import (
	"net/http"
	"web-forum/www/templates"
)

func HandleProfileSettings(stdWriter *http.ResponseWriter, stdRequest *http.Request) {
	infoToSend, account := Base(stdRequest, stdWriter)
	authorized := (*infoToSend)["Authorized"]
	defer templates.Index.Execute(*stdWriter, infoToSend)

	if !authorized.(bool) {
		(*infoToSend)["Title"] = "Нет доступа"
		templates.ContentAdd(infoToSend, templates.ProfileSettings, nil)
		return
	}

	(*infoToSend)["Title"] = "Настройки профиля"

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

	templates.ContentAdd(infoToSend, templates.ProfileSettings, contentToAdd)
}
