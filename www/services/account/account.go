package account

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"web-forum/system/db"
	"web-forum/system/rdb"
	jwt_token "web-forum/www/services/jwt-token"
)

type Account struct {
	Id       int
	Login    string
	Password string
	Username string
	Email    string
	IsAdmin  bool

	Sex         sql.NullString
	Avatar      sql.NullString
	Description sql.NullString
	SignText    sql.NullString

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type FastCacheStruct struct {
	Account *Account
	Time    time.Time
}

var FastCache = make(map[int]FastCacheStruct)

func ReadFromCookie(cookie *http.Cookie) (*Account, error) {
	var defaultAccount = Account{}
	tokenInfo, tokenErr := jwt_token.GetInfo(cookie.Value)

	if tokenErr != nil {
		return &defaultAccount, tokenErr
	}

	id, ok := tokenInfo["id"].(int64)

	if !ok {
		return &defaultAccount, fmt.Errorf("id not int")
	}

	accInfo, errGetAccount := GetById(int(id))

	if errGetAccount != nil {
		return &defaultAccount, errGetAccount
	}

	return accInfo, nil
}

func GetById(id int) (*Account, error) {
	veryFastCache, ok := FastCache[id]

	if ok {
		return veryFastCache.Account, nil
	}

	result, err := rdb.RedisDB.Get(ctx, fmt.Sprintf("aID:%d", id)).Result()

	if err == nil {
		outputAccount, errDeserialize := Deserialize(result)

		if errDeserialize != nil {
			return nil, errDeserialize
		}

		FastCache[id] = FastCacheStruct{
			Account: &outputAccount,
			Time:    time.Now().Add(time.Second * 10),
		}

		return &outputAccount, nil
	}

	accountInfo := &Account{}
	row := db.Postgres.QueryRow(ctx, "SELECT * FROM users WHERE id = $1;", id)

	if row == nil {
		return nil, fmt.Errorf("Account not found")
	}

	queryErr := row.Scan(
		&accountInfo.Id,
		&accountInfo.Login,
		&accountInfo.Password,
		&accountInfo.Username,
		&accountInfo.Email,
		&accountInfo.IsAdmin,
		&accountInfo.Sex,
		&accountInfo.Avatar,
		&accountInfo.Description,
		&accountInfo.SignText,
		&accountInfo.CreatedAt,
		&accountInfo.UpdatedAt,
	)

	if queryErr != nil {
		return nil, queryErr
	}

	rdb.RedisDB.Set(ctx, fmt.Sprintf("aID:%d", accountInfo.Id), accountInfo.Serialize(), time.Hour).Result()

	return accountInfo, nil
}

// GetByLogin
// Использовать только по крайней необходимости.
func GetByLogin(login string) (*Account, error) {
	accountInfo := &Account{}
	row := db.Postgres.QueryRow(ctx, "SELECT * FROM users WHERE login = $1;", login)

	if row == nil {
		return nil, fmt.Errorf("Account not found")
	}

	queryErr := row.Scan(
		&accountInfo.Id,
		&accountInfo.Login,
		&accountInfo.Password,
		&accountInfo.Username,
		&accountInfo.Email,
		&accountInfo.IsAdmin,
		&accountInfo.Sex,
		&accountInfo.Avatar,
		&accountInfo.Description,
		&accountInfo.SignText,
		&accountInfo.CreatedAt,
		&accountInfo.UpdatedAt,
	)

	if queryErr != nil {
		return nil, queryErr
	}

	return accountInfo, nil
}
