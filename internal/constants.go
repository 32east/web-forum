package internal

import (
	"database/sql"
	"os"
	"time"
)

const SiteName = "Форумчанский"
const LoginMinLength = 4
const PasswordMinLength = 8
const EmailMinLength = 4
const UsernameMinLength = 4
const AvatarsFilePath = "frontend/template/imgs/avatars/"
const AvatarsSize = 200.0
const MaxPaginatorTopics = 10
const HowMuchPagesWillBeVisibleInPaginator = 9 // Только нечётные числа!

var HmacSecret = []byte(os.Getenv("HMAC_SECRET"))

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

type Topic struct {
	Id         int
	ForumId    int
	Name       string
	Message    string
	Creator    int
	CreateTime time.Time
	UpdateTime sql.NullTime
}
