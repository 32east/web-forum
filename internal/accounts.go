package internal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
)

var CachedAccounts = make(map[string]Account)
var CachedAccountsById = make(map[int]*Account)

var db *sql.DB

func SetDB(dbSet *sql.DB) {
	db = dbSet
}

func ReadAccountFromCookie(cookie *http.Cookie, rdb *redis.Client) (*Account, error) {
	account := &Account{}
	login, err := rdb.Get(context.Background(), "AToken:"+cookie.Value).Result()

	if err != nil {
		return account, err
	}

	accInfo, errGetAccount := GetAccount(login)

	if errGetAccount != nil {
		return account, errGetAccount
	}

	return accInfo, nil
}

func GetAccountById(id int) (*Account, error) {
	val, ok := CachedAccountsById[id]

	if ok {
		return val, nil
	}

	accountInfo := &Account{}
	row := db.QueryRow("SELECT * FROM `users` WHERE `id` = ?;", id)

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
	row := db.QueryRow("SELECT * FROM `users` WHERE `login` =?;", login)

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
