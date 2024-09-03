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
const MaxPaginatorMessages = 10
const HowMuchPagesWillBeVisibleInPaginator = 9 // Только нечётные числа!!!

var HmacSecret = []byte(os.Getenv("HMAC_SECRET"))

type Topic struct {
	Id           int
	ForumId      int
	Name         string
	Creator      int
	CreateTime   time.Time
	UpdateTime   sql.NullTime
	MessageCount int
}

type Paginator struct {
	Objects     []interface{} // Здесь наши обрезанные объекты
	CurrentPage int           // Текущая страница
	AllPages    int           // Все страницы
	Err         error
}

type PaginatorArrows struct {
	Activated bool
	WhichPage int
}

type PaginatorConstructed struct {
	PagesArray []int
	Left       PaginatorArrows
	Right      PaginatorArrows
}
