package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/www/templates"
)

type LastRegisteredUser struct {
	Id       int
	Email    string
	Username string

	Sex       string
	Avatar    string
	CreatedAt string
}

type ChanCount struct {
	Key   string
	Value int
}

var lenCount = len("count:")

func redisGet(chanRdb chan ChanCount, key string) {
	conv := 0
	err := rdb.RedisDB.Get(ctx, key).Scan(&conv)
	key = key[lenCount:]

	if err != nil {
		fmtQuery := fmt.Sprintf("select count(*) from %s;", key)
		db.Postgres.QueryRow(ctx, fmtQuery).Scan(&conv)
	}

	chanRdb <- ChanCount{
		key, conv,
	}
}

func AdminMainPage(stdRequest *http.Request) {
	const errorFunction = "AdminMainPage"

	tx, err := db.Postgres.Begin(ctx)
	defer tx.Commit(ctx)

	if err != nil {
		templates.ContentAdd(stdRequest, templates.AdminMain, nil)

		return
	}

	chanRdb := make(chan ChanCount)
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

	var users []LastRegisteredUser
	rowsUser, err := tx.Query(ctx, `select id, username, email, sex, avatar, created_at from users as u order by id desc limit 10;`)

	if err != nil {
		system.ErrLog(errorFunction, err)
	} else {
		for rowsUser.Next() {
			var avatar, sex sql.NullString
			var registerDate time.Time
			var fmtSex string

			user := LastRegisteredUser{}
			scan := rowsUser.Scan(&user.Id, &user.Username, &user.Email, &sex, &avatar, &registerDate)

			if scan != nil {
				system.ErrLog(errorFunction, scan)
				continue
			}

			if avatar.Valid {
				user.Avatar = avatar.String
			}

			if sex.String == "m" {
				fmtSex = "Мужской"
			} else if sex.String == "f" {
				fmtSex = "Женский"
			}

			user.CreatedAt = registerDate.Format("2006-01-02 15:04:05")
			user.Sex = fmtSex

			users = append(users, user)
		}
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

	rows, err := db.Postgres.Query(ctx, `select * from forums order by id;`)

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
