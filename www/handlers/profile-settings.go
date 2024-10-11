package handlers

import (
	"net/http"
	account2 "web-forum/www/services/account"
	"web-forum/www/templates"
)

type sexSelect struct {
	Name       string
	Value      string
	IsSelected bool
}

func HandleProfileSettings(stdRequest *http.Request) {
	var reqCtx = stdRequest.Context()
	var account = reqCtx.Value("AccountData").(*account2.Account)
	var infoToSend = reqCtx.Value("InfoToSend").(map[string]interface{})
	var authorized = infoToSend["Authorized"]

	if !authorized.(bool) {
		infoToSend["Title"] = "Нет доступа"
		templates.ContentAdd(stdRequest, templates.ProfileSettings, nil)
		return
	}

	infoToSend["Title"] = "Настройки профиля"

	var contentToAdd = make(map[string]interface{})
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

	var sexStr = account.Sex.String
	var SexSelect []sexSelect

	SexSelect = append(SexSelect, sexSelect{
		Name:       "Мужской",
		Value:      "m",
		IsSelected: sexStr == "m",
	})

	SexSelect = append(SexSelect, sexSelect{
		Name:       "Женский",
		Value:      "f",
		IsSelected: sexStr == "f",
	})

	SexSelect = append(SexSelect, sexSelect{
		Name:       "Не указан",
		Value:      "",
		IsSelected: sexStr == "",
	})

	contentToAdd["SexSelect"] = SexSelect

	templates.ContentAdd(stdRequest, templates.ProfileSettings, contentToAdd)
}
