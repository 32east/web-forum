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
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/functions"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
	jwt_token "web-forum/internal/app/services/jwt-token"
)

var refreshTokenTime = 3600 * 72
var ctx = context.Background()

func CheckForSpecialCharacters(str string) bool {
	var countSpecialCharacters = strings.IndexFunc(str, func(r rune) bool {
		return r < 'A' || r > 'z'
	})
	return countSpecialCharacters > 0
}

func IsLoginAndPasswordLegalForActions(loginStr string, passwordStr string) (bool, string) {
	var success, reason = true, ""

	if loginStr == "" || passwordStr == "" {
		success, reason = false, "invalid login or password"
	}

	if CheckForSpecialCharacters(loginStr) || CheckForSpecialCharacters(passwordStr) {
		success, reason = false, "illegal characters"
	}

	return success, reason
}

func HandleRegister(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	var loginStr = reader.FormValue("login")
	var passwordStr = reader.FormValue("password")
	var username = reader.FormValue("username")
	var email = reader.FormValue("email")
	var success, reason = IsLoginAndPasswordLegalForActions(loginStr, passwordStr)

	if !success {
		answer["success"], answer["reason"] = false, reason
		return nil
	}

	tx, err := db.Postgres.Begin(ctx)

	if err != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return err
	}

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, `select
		case when login=$1 then 1 else 0 end,
       	case when email=$2 then 1 else 0 end,
       	case when username=$3 then 1 else 0 end
		from users as u1;`, loginStr, email, username)

	if err != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return err
	}

	var loginFounded, emailFounded, usernameFounded = 0, 0, 0

	for rows.Next() {
		errRow := rows.Scan(&loginFounded, &emailFounded, &usernameFounded)

		if errRow != nil {
			return errRow
		}
	}

	switch {
	case loginFounded == 1:
		answer["success"], answer["reason"] = false, "this login is already registered"
		return nil
	case emailFounded == 1:
		answer["success"], answer["reason"] = false, "this email is already registered"
		return nil
	case usernameFounded == 1:
		answer["success"], answer["reason"] = false, "this username is already registered"
		return nil
	}

	var loginLen = functions.Utf8Length(loginStr)
	var passwordLen = functions.Utf8Length(passwordStr)
	var usernameLen = functions.Utf8Length(username)
	var emailLen = functions.Utf8Length(email)

	switch {
	case loginLen < models.LoginMinLength:
		answer["success"], answer["reason"] = false, "login too short"
		return nil
	case passwordLen < models.PasswordMinLength:
		answer["success"], answer["reason"] = false, "password too short"
		return nil
	case usernameLen < models.UsernameMinLength:
		answer["success"], answer["reason"] = false, "username too short"
		return nil
	case loginLen > models.LoginMaxLength:
		answer["success"], answer["reason"] = false, "login too long"
		return nil
	case passwordLen > models.PasswordMaxLength:
		answer["success"], answer["reason"] = false, "password too long"
		return nil
	case usernameLen > models.UsernameMaxLength:
		answer["success"], answer["reason"] = false, "username too long"
		return nil
	case emailLen < models.EmailMinLength:
		answer["success"], answer["reason"] = false, "email too short"
		return nil
	case emailLen > models.EmailMaxLength:
		answer["success"], answer["reason"] = false, "username too long"
		return nil
	default:
	}

	if _, err := mail.ParseAddress(email); err != nil {
		answer["success"], answer["reason"] = false, "invalid email"
		return nil
	}

	newSha256Writer := sha256.New()
	newSha256Writer.Write([]byte(passwordStr))
	hexPassword := hex.EncodeToString(newSha256Writer.Sum(nil))

	if _, execErr := tx.Exec(ctx, "insert into users(login, password, username, email, created_at) values ($1, $2, $3, $4, CURRENT_TIMESTAMP);",
		loginStr,
		hexPassword,
		username,
		email); execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	rdb.RedisDB.Do(ctx, "incrby", "count:users", 1)

	answer["success"] = true
	return nil
}

// TODO: Могут насрать запросами, что по итогу выльется в DDOS.
func HandleLogin(writer http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	var loginStr = reader.FormValue("login")
	var passwordStr = reader.FormValue("password")
	var success, reason = IsLoginAndPasswordLegalForActions(loginStr, passwordStr)

	if !success {
		answer["success"], answer["reason"] = false, reason
		return nil
	}

	var accountInfo, queryErr = account.GetByLogin(loginStr)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, "account not founded"
		return nil
	}

	var toSha256 = sha256.New()
	toSha256.Write([]byte(passwordStr))
	var hexPassword = hex.EncodeToString(toSha256.Sum(nil))

	if accountInfo.Password != hexPassword {
		answer["success"], answer["reason"] = false, "account not founded"
		return nil
	}

	var accessToken, errAccess = jwt_token.GenerateNew(accountInfo.Id, "access")

	if errAccess != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return errAccess
	}

	var refreshToken, errRefresh = jwt_token.GenerateNew(accountInfo.Id, "refresh")

	if errRefresh != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return errRefresh
	}

	answer["success"], answer["access_token"], answer["refresh_token"] = true, accessToken, refreshToken
	answer["access_token_exp"], answer["refresh_token_exp"] = 3600*12, refreshTokenTime

	http.SetCookie(writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		// HttpOnly: true,
		// В фронтенде у меня обновление токена идёт через JS, но поскольку я JS плохо знаю я хз как без JS обновление сделать.
		// Так что... да. Здесь у меня просто выбора нет.
	})

	return nil
}

func HandleRefreshToken(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var mapToken = map[string]string{}
	var jsonErr = json.NewDecoder(r.Body).Decode(&mapToken)

	if jsonErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return jsonErr
	}

	var refreshToken = mapToken["refresh_token"]

	if refreshToken == "" {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return nil
	}

	var tokenClaim, err = jwt_token.GetInfo(refreshToken)

	if err != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return err
	}

	var accountId = tokenClaim["id"]
	var _, errGetAccount = account.GetById(int(accountId.(int64)))

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return nil
	}

	var accessToken, errAccess = jwt_token.GenerateNew(accountId.(int64), "access")

	if errAccess != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return errAccess
	}

	var newRefreshToken, errRefresh = jwt_token.GenerateNew(accountId.(int64), "refresh")

	if errRefresh != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return errRefresh
	}

	var tx, txErr = db.Postgres.Begin(ctx)

	if txErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return txErr
	}

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	var _, execErr = tx.Exec(ctx, "delete from tokens where refresh_token = $1;", refreshToken)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	var fmtQuery = fmt.Sprintf("insert into tokens (account_id, refresh_token, expiresat) values ($1, $2, now() + interval '%d second');", refreshTokenTime)
	_, execErr = tx.Exec(ctx, fmtQuery, accountId, newRefreshToken)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	answer["success"], answer["access_token"], answer["refresh_token"] = true, accessToken, newRefreshToken

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		// HttpOnly: true,
	})

	return nil
}

func HandleLogout(writer http.ResponseWriter, _ *http.Request, answer map[string]interface{}) error {
	http.SetCookie(writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		// HttpOnly: true,
	})

	answer["success"] = true
	return nil
}
