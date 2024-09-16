package handlers

import (
	"database/sql"
	"net/http"
	"time"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
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

func AdminMainPage(stdRequest *http.Request) {
	const errorFunction = "AdminMainPage"

	tx, err := db.Postgres.Begin(ctx)
	defer tx.Commit(ctx)

	if err != nil {
		templates.ContentAdd(stdRequest, templates.AdminMain, nil)

		return
	}

	rows, err := tx.Query(ctx, `select (select count(*) from topics),
		(select count(*) from messages),
		(select count(*) from users);`)

	var topicsCount, messagesCount, usersCount int

	if err != nil {
		system.ErrLog(errorFunction, err)
	} else {
		for rows.Next() {
			err = rows.Scan(&topicsCount, &messagesCount, &usersCount)

			if err != nil {
				system.ErrLog(errorFunction, err)
				break
			}
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
		"TopicsCount":         topicsCount,
		"MessagesCount":       messagesCount,
		"UsersCount":          usersCount,
		"LastRegisteredUsers": users,
	}

	templates.ContentAdd(stdRequest, templates.AdminMain, contentToAdd)
}

func AdminCategoriesPage(stdRequest *http.Request) {
	const errorFunction = "AdminCategoriesPage"

	rows, err := db.Postgres.Query(ctx, `select * from forums;`)

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

func AdminCategoryCreate(stdRequest *http.Request) {
	templates.ContentAdd(stdRequest, templates.AdminCreateCategory, nil)
}
