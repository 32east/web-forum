package internal

import (
	"database/sql"
	"os"
	"time"
)

const SiteName = "Форумчанский"

const LoginMinLength = 4
const LoginMaxLength = 32

const PasswordMinLength = 8
const PasswordMaxLength = 64

const EmailMinLength = 4
const EmailMaxLength = 64

const UsernameMinLength = 4
const UsernameMaxLength = 24

const AvatarsFilePath = "frontend/imgs/avatars/"
const AvatarsSize = 256.0
const MaxPaginatorMessages = 10
const HowMuchPagesWillBeVisibleInPaginator = 9 // Только нечётные числа!!!

var HmacSecret = []byte(os.Getenv("HMAC_SECRET"))

type Category struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TopicsCount int
}

type Topic struct {
	Id           int
	ForumId      int
	Name         string
	Creator      int
	CreateTime   time.Time
	UpdateTime   sql.NullTime
	MessageCount int
}

type Message struct {
	Id         int
	TopicId    int
	CreatorId  int
	Message    string
	CreateTime time.Time
	UpdateTime sql.NullTime
}

type ProfileMessage struct {
	TopicId    int
	TopicName  string
	Message    string
	CreateTime string
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
