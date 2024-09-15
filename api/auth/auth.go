package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	initialize_functions "web-forum/www/initialize-functions"
	"web-forum/www/services/account"
	jwt_token "web-forum/www/services/jwt-token"
)

var ctx = context.Background()

func CheckForSpecialCharacters(str string) bool {
	countSpecialCharacters := strings.IndexFunc(str, func(r rune) bool {
		return r < 'A' || r > 'z'
	})

	return countSpecialCharacters > 0
}

func PrepareHandle(writer *http.ResponseWriter) (*json.Encoder, map[string]interface{}) {
	header := (*writer).Header()
	header.Add("content-type", "application/json")

	newJSONEncoder := json.NewEncoder(*writer)
	answer := make(map[string]interface{})

	return newJSONEncoder, answer
}

func IsLoginAndPasswordLegalForActions(loginStr string, passwordStr string) (bool, string) {
	success, reason := true, ""

	if loginStr == "" || passwordStr == "" {
		success, reason = false, "invalid login or password"
	}

	if CheckForSpecialCharacters(loginStr) || CheckForSpecialCharacters(passwordStr) {
		success, reason = false, "illegal characters"
	}

	return success, reason
}

func HandleRegister(writer *http.ResponseWriter, reader *http.Request) {
	const errFunction = "HandleRegister"

	newJSONEncoder, answer := PrepareHandle(writer)
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

	loginStr := reader.FormValue("login")
	passwordStr := reader.FormValue("password")
	username := reader.FormValue("username")
	email := reader.FormValue("email")

	success, reason := IsLoginAndPasswordLegalForActions(loginStr, passwordStr)

	if !success {
		answer["success"], answer["reason"] = false, reason
		return
	}

	rows, err := db.Postgres.Query(ctx, `select
		case when login=$1 then 1 else 0 end,
       	case when email=$2 then 1 else 0 end,
       	case when username=$3 then 1 else 0 end
		from users as u1;`, loginStr, email, username)

	if err != nil {
		answer["success"], answer["reason"] = false, err.Error()
		return
	}

	loginFounded, emailFounded, usernameFounded := 0, 0, 0
	errReason := ""

	for rows.Next() {
		errRow := rows.Scan(&loginFounded, &emailFounded, &usernameFounded)

		if errRow != nil {
			errReason = errRow.Error()
			break
		}
	}

	switch {
	case errReason != "":
		answer["success"], answer["reason"] = false, "internal server error"
		return
	case loginFounded == 1:
		answer["success"], answer["reason"] = false, "this login is already registered"
		return
	case emailFounded == 1:
		answer["success"], answer["reason"] = false, "this email is already registered"
		return
	case usernameFounded == 1:
		answer["success"], answer["reason"] = false, "this username is already registered"
		return
	}

	loginLen := internal.Utf8Length(loginStr)
	passwordLen := internal.Utf8Length(passwordStr)
	usernameLen := internal.Utf8Length(username)
	emailLen := internal.Utf8Length(email)

	switch {
	case loginLen < internal.LoginMinLength:
		answer["success"], answer["reason"] = false, "login too short"
		return
	case passwordLen < internal.PasswordMinLength:
		answer["success"], answer["reason"] = false, "password too short"
		return
	case usernameLen < internal.UsernameMinLength:
		answer["success"], answer["reason"] = false, "username too short"
		return
	case loginLen > internal.LoginMaxLength:
		answer["success"], answer["reason"] = false, "login too long"
		return
	case passwordLen > internal.PasswordMaxLength:
		answer["success"], answer["reason"] = false, "password too long"
		return
	case usernameLen > internal.UsernameMaxLength:
		answer["success"], answer["reason"] = false, "username too long"
		return
	case emailLen < internal.EmailMinLength:
		answer["success"], answer["reason"] = false, "email too short"
		return
	case emailLen > internal.EmailMaxLength:
		answer["success"], answer["reason"] = false, "username too long"
		return
	default:
	}

	if _, err := mail.ParseAddress(email); err != nil {
		answer["success"], answer["reason"] = false, "invalid email"
		return
	}

	newSha256Writer := sha256.New()
	newSha256Writer.Write([]byte(passwordStr))
	hexPassword := hex.EncodeToString(newSha256Writer.Sum(nil))

	row := db.Postgres.QueryRow(ctx, "INSERT INTO users(login, password, username, email, created_at) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP) returning id;",
		loginStr,
		hexPassword,
		username,
		email)

	var accountId int
	idErr := row.Scan(&accountId)

	if idErr != nil {
		answer["success"], answer["reason"] = false, idErr.Error()
		return
	}

	initialize_functions.CreateProfilePage(accountId)

	answer["success"] = true
}

// TODO: Могут насрать запросами, что по итогу выльется в DDOS.
func HandleLogin(writer *http.ResponseWriter, reader *http.Request) {
	const errFunction = "HandleLogin"
	newJSONEncoder, answer := PrepareHandle(writer)

	defer func() {
		if !answer["success"].(bool) {
			system.ErrLog(errFunction, fmt.Errorf(string(reader.RemoteAddr)+" > "+answer["reason"].(string)))
		}
	}()

	defer newJSONEncoder.Encode(answer)

	if reader.Method != "POST" {
		answer["success"], answer["reason"] = false, "method not allowed"

		return
	}

	loginStr := reader.FormValue("login")
	passwordStr := reader.FormValue("password")

	success, reason := IsLoginAndPasswordLegalForActions(loginStr, passwordStr)

	if !success {
		answer["success"], answer["reason"] = false, reason

		return
	}

	accountInfo, queryErr := account.GetByLogin(loginStr)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()+" in query"

		return
	}

	toSha256 := sha256.New()
	toSha256.Write([]byte(passwordStr))
	hexPassword := hex.EncodeToString(toSha256.Sum(nil))

	if accountInfo.Password != hexPassword {
		answer["success"], answer["reason"] = false, "wrong password"

		return
	}

	accessToken, errAccess := jwt_token.GenerateNew(accountInfo.Id, "access")

	if errAccess != nil {
		answer["success"], answer["reason"] = false, errAccess.Error()

		return
	}

	refreshToken, errRefresh := jwt_token.GenerateNew(accountInfo.Id, "refresh")

	if errRefresh != nil {
		answer["success"], answer["reason"] = false, errRefresh

		return
	}

	errRTokenSet := rdb.RedisDB.Set(ctx, "RToken:"+refreshToken, loginStr, time.Hour*72)

	if errRTokenSet.Err() != nil {
		answer["success"], answer["reason"] = false, errRTokenSet.Err().Error()
		return
	}

	fmt.Println("Заносим в кэш: " + loginStr)

	answer["success"], answer["access_token"], answer["refresh_token"] = true, accessToken, refreshToken

	// time.Hour * 12, time.Hour * 72
	answer["access_token_exp"], answer["refresh_token_exp"] = 3600*12, 3600*72

	http.SetCookie(*writer, &http.Cookie{
		Name:    "access_token",
		Value:   accessToken,
		Expires: time.Now().Add(time.Hour * 12),
		Path:    "/",
	})

	http.SetCookie(*writer, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 72),
	})
}

func HandleLogout(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := PrepareHandle(writer)
	defer newJSONEncoder.Encode(answer)

	if reader.Method != "POST" {
		answer["success"], answer["reason"] = false, "method not allowed"

		return
	}

	http.SetCookie(*writer, &http.Cookie{
		Name:   "access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.SetCookie(*writer, &http.Cookie{
		Name:   "refresh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
		// HttpOnly: true,
	})

	answer["success"] = true
}
