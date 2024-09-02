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
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"web-forum/internal"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func CheckForSpecialCharacters(str string) bool {
	countSpecialCharacters := strings.IndexFunc(str, func(r rune) bool {
		return r < 'A' || r > 'z'
	})

	return countSpecialCharacters > 0
}

func GenerateNewJWTToken(login string, additional_param string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":   login,
		"additional": additional_param,
		"rand":       randomBytes,
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

func CheckValueInDatabase(db *sql.DB, sendError chan error, key string, value string) {
	existsValue := ""

	queryRow := db.QueryRow("SELECT ? FROM `users` WHERE ? = ?;", key, key, value)
	err := queryRow.Scan(&existsValue)

	if existsValue != "" {
		sendError <- fmt.Errorf("account with same %s is already exists", key)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		sendError <- err
	}

	sendError <- nil
}

func HandleRegister(writer *http.ResponseWriter, reader *http.Request, db *sql.DB) {
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

	go CheckValueInDatabase(db, receiveError, "login", loginStr)
	go CheckValueInDatabase(db, receiveError, "email", email)
	go CheckValueInDatabase(db, receiveError, "username", username)

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

	_, dbErr := db.Exec("REPLACE INTO `users` (login, password, username, email, created_at) VALUES (?, ?, ?, ?, ?);",
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

func HandleLogin(writer *http.ResponseWriter, reader *http.Request, db *sql.DB, rdb *redis.Client) {
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

	rdbGet := rdb.Get(ctx, loginStr)
	val, errRdbGet := rdbGet.Result()

	// Эта проверка возможно никогда не будет срабатывать.
	if errRdbGet == nil {
		buffer := map[string]string{}
		unmarshalledValue := json.Unmarshal([]byte(val), &buffer)

		if unmarshalledValue != nil {
			// Мы в ключе Login храним не JSON структуру??
			answer["success"], answer["reason"] = false, unmarshalledValue.Error()
			return
		}

		accessToken := buffer["access_token"]
		answer["success"], answer["access_token"], answer["refresh_token"] = true, accessToken, buffer["refresh_token"]

		// Бля ну я хз как тут узнать, когда у нас срок годности у токена истекает.
		http.SetCookie(*writer, &http.Cookie{
			Name:  "access_token",
			Value: accessToken,
			Path:  "/",
		})

		http.SetCookie(*writer, &http.Cookie{
			Name:  "refresh_token",
			Value: buffer["refresh_token"],
			Path:  "/",
		})

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

	errATokenSet := rdb.Set(ctx, "AToken:"+accessToken, loginStr, time.Hour*12)

	if errATokenSet.Err() != nil {
		// Если что можно будет добавить обработчик в базу данных.
		// Но это уже потом...

		answer["success"], answer["reason"] = false, errATokenSet.Err().Error()
		return
	}

	errRTokenSet := rdb.Set(ctx, "RToken:"+refreshToken, loginStr, time.Hour*72)

	if errRTokenSet.Err() != nil {
		answer["success"], answer["reason"] = false, errATokenSet.Err().Error()
		return
	}

	// Будет служить анти-спамом.
	// То есть если человек попросит токен в эти 60 секунд,
	// то ему выдастся уже кэшированный токен.
	marshal, marshalErr := json.Marshal(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

	if marshalErr != nil {
		answer["success"], answer["reason"] = false, marshalErr.Error()

		return
	}

	setErr := rdb.Set(ctx, loginStr, string(marshal), time.Hour*12)
	redisQueryErr := setErr.Err()

	if redisQueryErr != nil {
		answer["success"], answer["reason"] = false, setErr

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

func HandleLogout(writer *http.ResponseWriter, reader *http.Request, db *sql.DB, rdb *redis.Client) {
	newJSONEncoder, answer := PrepareHandle(writer)
	defer newJSONEncoder.Encode(answer)

	if reader.Method != "POST" {
		answer["success"], answer["reason"] = false, "method not allowed"

		return
	}

	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "invalid jwt token"

		return
	}

	loginStr, errRdbGet := rdb.Get(ctx, "AToken:"+cookie.Value).Result()

	if errRdbGet != nil {
		answer["success"], answer["reason"] = false, "invalid atoken"

		return
	}

	mapWithAccessAndRefreshToken, mapRdbErr := rdb.Get(ctx, loginStr).Result()

	if mapRdbErr != nil {
		answer["success"], answer["reason"] = false, mapRdbErr.Error()

		return
	}

	buffer := map[string]string{}
	unmarshalErr := json.Unmarshal([]byte(mapWithAccessAndRefreshToken), &buffer)

	if unmarshalErr != nil {
		// Мы в ключе Login храним не JSON структуру??
		answer["success"], answer["reason"] = false, unmarshalErr.Error()
		return
	}

	findAccount, errToFind := account.GetAccount(loginStr)

	if errToFind != nil {
		answer["success"], answer["reason"] = false, "couldn't find account"
		return
	}

	delete(account.CachedAccounts, loginStr)
	delete(account.CachedAccountsById, findAccount.Id)

	rdb.Del(ctx, loginStr)
	rdb.Del(ctx, "AToken:"+buffer["access_token"])
	rdb.Del(ctx, "RToken:"+buffer["refresh_token"])

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
	})

	answer["success"] = true
}
