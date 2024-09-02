package web

import (
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"web-forum/internal"
)

// TODO: Позже сделать обработчик, что если access_token уже устаревший, то гляди на refresh_token.
// TODO: Если и refresh_token устаревший, то всё пизда.

func HandleBase(stdRequest *http.Request, writer *http.ResponseWriter, rdb *redis.Client) (*map[string]interface{}, *internal.Account) {
	go TokensRefreshInRedis(stdRequest, writer, rdb)

	infoToSend := make(map[string]interface{})
	cookie, err := stdRequest.Cookie("access_token")

	infoToSend["Authorized"] = false

	account := &internal.Account{}

	if err != nil {
		return &infoToSend, account
	}

	account, errGetAccount := internal.ReadAccountFromCookie(cookie, rdb)

	if errGetAccount != nil {
		return &infoToSend, account
	}

	infoToSend["Authorized"] = true
	infoToSend["Username"] = account.Username

	if account.Avatar.Valid {
		infoToSend["Avatar"] = account.Avatar.String
	}

	return &infoToSend, account
}

func HandleMainPage(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = internal.SiteName
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	categorys, err := GetForums(db)

	if err != nil {
		log.Fatal(err)
	}

	ContentAdd(infoToSend, ForumTemplate, map[string]interface{}{
		"categorys":          categorys,
		"categorys_is_empty": len(*categorys) == 0,
	})
}

func HandleLoginPage(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = "Авторизация"
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	ContentAdd(infoToSend, LoginTemplate, nil)
}

func HandleRegisterPage(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = "Регистрация"
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	ContentAdd(infoToSend, RegisterTemplate, nil)
}

func HandleProfileSettings(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, account := HandleBase(stdRequest, stdWriter, rdb)
	authorized := (*infoToSend)["Authorized"]
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	if !authorized.(bool) {
		(*infoToSend)["Title"] = "Нет доступа"
		ContentAdd(infoToSend, ProfileSettingsTemplate, nil)
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

	ContentAdd(infoToSend, ProfileSettingsTemplate, contentToAdd)
}

func HandleFAQPage(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = "FAQ"
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	ContentAdd(infoToSend, FAQTemplate, nil)
}

func HandleUsersPage(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = "Юзеры"
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	ContentAdd(infoToSend, UsersTemplate, nil)
}

func HandleTopicCreate(stdWriter *http.ResponseWriter, stdRequest *http.Request, db *sql.DB, rdb *redis.Client) {
	infoToSend, _ := HandleBase(stdRequest, stdWriter, rdb)
	(*infoToSend)["Title"] = "Создание нового топика"
	defer IndexTemplate.Execute(*stdWriter, infoToSend)

	forums, err := GetForums(db)

	if err != nil {
		panic(err)
	}

	var categorys []interface{}
	currentCategory := stdRequest.FormValue("category")

	for _, output := range *forums {
		outputToMap := output.(map[string]interface{})
		forumId := outputToMap["forum_id"]

		categorys = append(categorys, map[string]interface{}{
			"forum_name":  outputToMap["forum_name"].(string),
			"forum_id":    outputToMap["forum_id"].(int),
			"is_selected": fmt.Sprint(forumId) == currentCategory,
		})
	}

	ContentAdd(infoToSend, CreateNewTopicTemplate, map[string]interface{}{
		"categorys": categorys,
	})
}
