package handlers

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
	"time"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/www/services/paginator"
	"web-forum/www/templates"
)

type chanCount struct {
	Key   string
	Value int
}

type AdmAccount struct {
	Id          int
	Username    string
	Email       string
	IsAdmin     bool
	Sex         string
	Avatar      string
	Description string
	SignText    string
	CreatedAt   string
	UpdatedAt   string
}

var lenCount = len("count:")

func redisGet(chanRdb chan chanCount, key string) {
	conv := 0
	err := rdb.RedisDB.Get(ctx, key).Scan(&conv)
	key = key[lenCount:]

	if err != nil {
		fmtQuery := fmt.Sprintf("select count(*) from %s;", key)
		db.Postgres.QueryRow(ctx, fmtQuery).Scan(&conv)
	}

	chanRdb <- chanCount{
		key, conv,
	}
}

func addToMap(rows *pgx.Rows, sendInfo *[]AdmAccount) {
	for (*rows).Next() {
		var sex, avatar, description, signText sql.NullString
		var createdAt time.Time
		var updatedAt sql.NullTime

		account := AdmAccount{}
		scanErr := (*rows).Scan(&account.Id, &account.Username, &account.Email, &account.IsAdmin, &sex, &avatar, &description, &signText, &createdAt, &updatedAt)

		if scanErr != nil {
			system.ErrLog("admin.addToMap", scanErr)
		}

		if sex.Valid {
			account.Sex = sex.String
		}

		if avatar.Valid {
			account.Avatar = avatar.String
		}

		if description.Valid {
			account.Description = description.String
		}

		if signText.Valid {
			account.SignText = signText.String
		}

		account.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		if updatedAt.Valid {
			account.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
		}

		*sendInfo = append(*sendInfo, account)
	}
}

func AdminMainPage(stdRequest *http.Request) {
	const errorFunction = "AdminMainPage"

	tx, err := db.Postgres.Begin(ctx)

	if tx != nil {
		defer tx.Commit(ctx)
	}

	if err != nil {
		templates.ContentAdd(stdRequest, templates.AdminMain, nil)

		return
	}

	chanRdb := make(chan chanCount)
	defer close(chanRdb)

	var countsInfo = map[string]int{}

	for _, val := range []string{"topics", "messages", "users"} {
		go redisGet(chanRdb, "count:"+val)
	}

	count := 3

	for {
		val := <-chanRdb
		countsInfo[val.Key] = val.Value
		count -= 1

		if count <= 0 {
			break
		}
	}

	var users []AdmAccount
	rowsUser, err := tx.Query(ctx, `select id, username, email, is_admin, sex, avatar, description, sign_text, created_at, updated_at from users as u order by id desc limit 10;`)

	if err != nil {
		system.ErrLog(errorFunction, err)
	} else {
		addToMap(&rowsUser, &users)
	}

	contentToAdd := map[string]interface{}{
		"TopicsCount":         countsInfo["topics"],
		"MessagesCount":       countsInfo["messages"],
		"UsersCount":          countsInfo["users"],
		"LastRegisteredUsers": users,
	}

	templates.ContentAdd(stdRequest, templates.AdminMain, contentToAdd)
}

func AdminCategoriesPage(stdRequest *http.Request) {
	const errorFunction = "AdminCategoriesPage"

	rows, err := db.Postgres.Query(ctx, `select * from categorys order by id;`)

	if err != nil {
		system.ErrLog(errorFunction, err)
	}

	var categories []internal.Category

	for rows.Next() {
		category := internal.Category{}
		scanErr := rows.Scan(&category.Id, &category.Name, &category.Description, &category.TopicsCount)

		if scanErr != nil {
			system.ErrLog(errorFunction, scanErr)
			continue
		}

		categories = append(categories, category)
	}

	templates.ContentAdd(stdRequest, templates.AdminCategories, categories)
}

func AdminUsersPage(r *http.Request) {
	const errorFunction = "AdminUsersPage"

	page := 1
	pageStr := r.FormValue("page")

	if pageStr != "" {
		pageNum, err := strconv.Atoi(pageStr)

		if err != nil {
			system.ErrLog(errorFunction, err)
		}

		page = pageNum
	}

	search := r.FormValue("search")

	preQuery := internal.PaginatorPreQuery{
		TableName:     "users",
		OutputColumns: "id, username, email, is_admin, sex, avatar, description, sign_text, created_at, updated_at",
		Page:          page,
	}

	if search != "" {
		// Это будет медленно.
		var queryCount int
		db.Postgres.QueryRow(ctx, "select count(*) from users where username like $1", "%"+search+"%").Scan(&queryCount)

		preQuery.WhereColumn = "username"
		preQuery.WhereOperator = "like"
		preQuery.WhereValue = "%" + search + "%"
	} else {
		usersCount, usersCountErr := rdb.RedisDB.Get(ctx, "count:users").Result()

		if usersCountErr != nil {
			system.ErrLog(errorFunction, usersCountErr)

			usersCount = "0"
		}

		conv, err := strconv.Atoi(usersCount)

		if err != nil {
			system.ErrLog(errorFunction, err)
			conv = 0
		}

		preQuery.QueryCount.PreparedValue = conv
	}

	tx, rows, _, err := paginator.Query(preQuery)

	if tx != nil {
		defer tx.Commit(ctx)
	}

	if err != nil {
		system.ErrLog(errorFunction, err)
	}

	var sendInfo []AdmAccount

	addToMap(&rows, &sendInfo)

	contentToAdd := map[string]interface{}{
		"Users": sendInfo,
	}

	templates.ContentAdd(r, templates.AdminUsers, contentToAdd)
}
