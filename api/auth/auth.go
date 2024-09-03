package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"web-forum/internal"
	"web-forum/system/redisDb"
	"web-forum/system/sqlDb"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func CheckForSpecialCharacters(str string) bool {
	countSpecialCharacters := strings.IndexFunc(str, func(r rune) bool {
		return r < 'A' || r > 'z'
	})

	return countSpecialCharacters > 0
}

func GenerateNewJWTToken(login string, additionalParam string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(time.Hour * 24)

	if additionalParam == "refresh" {
		expirationTime = time.Now().Add(time.Hour * 72)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":      login,
		"additional": additionalParam,
		"expiresAt":  fmt.Sprintf("%d", expirationTime.Unix()),
	})

	tokenStr, err := token.SignedString(internal.HmacSecret)

	return tokenStr, err
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

func CheckValueInDatabase(tx *sql.Tx, sendError chan error, key string, value string) {
	existsValue := ""

	queryRow := tx.QueryRow("SELECT ? FROM `users` WHERE ? = ?;", key, key, value)
	err := queryRow.Scan(&existsValue)

	if existsValue != "" {
		sendError <- fmt.Errorf("account with same %s is already exists", key)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		sendError <- err
	}

	sendError <- nil
}

func HandleRegister(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := PrepareHandle(writer)
	defer newJSONEncoder.Encode(answer)

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

	receiveError := make(chan error)

	txToCheck, errTx := sqlDb.MySqlDB.Begin()
	defer txToCheck.Commit()

	if errTx != nil {
		answer["success"], answer["reason"] = false, errTx.Error()

		return
	}

	go CheckValueInDatabase(txToCheck, receiveError, "login", loginStr)
	go CheckValueInDatabase(txToCheck, receiveError, "email", email)
	go CheckValueInDatabase(txToCheck, receiveError, "username", username)

	workersCount := 3

	for errDb := range receiveError {
		workersCount -= 1

		if workersCount <= 0 {
			close(receiveError)
		}

		if errDb != nil {
			answer["success"], answer["reason"] = false, errDb.Error()
			close(receiveError)

			return
		}
	}

	switch {
	case len(loginStr) < internal.LoginMinLength:
		answer["success"], answer["reason"] = false, "login is too short"
		return
	case len(passwordStr) < internal.PasswordMinLength:
		answer["success"], answer["reason"] = false, "password too short"
		return
	case len(username) < internal.UsernameMinLength:
		answer["success"], answer["reason"] = false, "username too short"
		return
	case len(email) < internal.EmailMinLength:
		answer["success"], answer["reason"] = false, "email too short"
		return
	case username == "settings":
		answer["success"], answer["reason"] = false, "invalid username"
	}

	if _, err := mail.ParseAddress(email); err != nil {
		answer["success"], answer["reason"] = false, "invalid email"

		return
	}

	newSha256Writer := sha256.New()
	newSha256Writer.Write([]byte(passwordStr))
	hexPassword := hex.EncodeToString(newSha256Writer.Sum(nil))

	accountCreateTime := time.Now()

	_, dbErr := sqlDb.MySqlDB.Exec("REPLACE INTO `users` (login, password, username, email, created_at) VALUES (?, ?, ?, ?, ?);",
		loginStr,
		hexPassword,
		username,
		email,
		accountCreateTime)

	if dbErr != nil {
		answer["success"], answer["reason"] = false, dbErr.Error()

		return
	}

	answer["success"] = true
}

func HandleLogin(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := PrepareHandle(writer)

	defer func() {
		if !answer["success"].(bool) {
			log.Println(string(reader.RemoteAddr) + " > " + answer["reason"].(string))
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

	accountInfo, queryErr := account.GetAccount(loginStr)

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

	accessToken, errAccess := GenerateNewJWTToken(loginStr, "access")

	if errAccess != nil {
		answer["success"], answer["reason"] = false, errAccess.Error()

		return
	}

	refreshToken, errRefresh := GenerateNewJWTToken(loginStr, "refresh")

	if errRefresh != nil {
		answer["success"], answer["reason"] = false, errRefresh

		return
	}

	errRTokenSet := redisDb.RedisDB.Set(ctx, "RToken:"+refreshToken, loginStr, time.Hour*72)

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
	//
	//cookie, err := reader.Cookie("access_token")
	//
	//if err != nil {
	//	answer["success"], answer["reason"] = false, "invalid jwt token"
	//
	//	return
	//}
	//
	//_, tokenErr := GetTokenInfo(cookie.Value)
	//
	//if tokenErr != nil {
	//	answer["success"], answer["reason"] = false, "couldn't find account"
	//	return
	//}

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
