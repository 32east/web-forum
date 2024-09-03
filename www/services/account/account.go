package account

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"web-forum/system/sqlDb"
	jwt_token "web-forum/www/services/jwt-token"
)

type Account struct {
	Id       int
	Login    string
	Password string
	Username string
	Email    string

	Avatar      sql.NullString
	Description sql.NullString
	SignText    sql.NullString

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

var CachedAccounts = make(map[string]Account)
var CachedAccountsById = make(map[int]*Account)

func ReadAccountFromCookie(cookie *http.Cookie) (*Account, error) {
	tokenInfo, tokenErr := jwt_token.GetTokenInfo(cookie.Value)

	if tokenErr != nil {
		return nil, tokenErr
	}

	accInfo, errGetAccount := GetAccount(tokenInfo["login"].(string))

	if errGetAccount != nil {
		return nil, errGetAccount
	}

	return accInfo, nil
}

func GetAccountById(id int) (*Account, error) {
	val, ok := CachedAccountsById[id]

	if ok {
		return val, nil
	}

	accountInfo := &Account{}
	row := sqlDb.MySqlDB.QueryRow("SELECT * FROM `users` WHERE `id` = ?;", id)

	if row == nil {
		return accountInfo, fmt.Errorf("Account not found")
	}

	queryErr := row.Scan(
		&accountInfo.Id,
		&accountInfo.Login,
		&accountInfo.Password,
		&accountInfo.Username,
		&accountInfo.Email,
		&accountInfo.Avatar,
		&accountInfo.Description,
		&accountInfo.SignText,
		&accountInfo.CreatedAt,
		&accountInfo.UpdatedAt,
	)

	if queryErr != nil {
		return accountInfo, queryErr
	}

	CachedAccounts[accountInfo.Login] = *accountInfo

	valId, _ := CachedAccounts[accountInfo.Login]
	CachedAccountsById[accountInfo.Id] = &valId

	return accountInfo, nil
}

func GetAccount(login string) (*Account, error) {
	val, ok := CachedAccounts[login]

	if ok {
		return &val, nil
	}

	accountInfo := &Account{}
	row := sqlDb.MySqlDB.QueryRow("SELECT * FROM `users` WHERE `login` =?;", login)

	if row == nil {
		return accountInfo, fmt.Errorf("Account not found")
	}

	queryErr := row.Scan(
		&accountInfo.Id,
		&accountInfo.Login,
		&accountInfo.Password,
		&accountInfo.Username,
		&accountInfo.Email,
		&accountInfo.Avatar,
		&accountInfo.Description,
		&accountInfo.SignText,
		&accountInfo.CreatedAt,
		&accountInfo.UpdatedAt,
	)

	if queryErr != nil {
		return accountInfo, queryErr
	}

	CachedAccounts[login] = *accountInfo
	CachedAccountsById[accountInfo.Id] = accountInfo

	return accountInfo, nil
}
