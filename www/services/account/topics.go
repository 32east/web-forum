package account

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5"
	"web-forum/system"
)

type UserTopic struct {
	Username string
	Avatar   sql.NullString
	SignText sql.NullString
}

var ctx = context.Background()

func GetFromSlice(tempUsers []int, tx pgx.Tx) map[int]UserTopic {
	const errorFunction = "topics.GetFromSlice"
	var rowsUsers, errUsers = tx.Query(ctx, `select id, username, avatar, sign_text
			from users
			where id = any($1);`, tempUsers)

	var usersInfo = map[int]UserTopic{}

	if errUsers != nil {
		system.ErrLog(errorFunction, errUsers)
		return usersInfo
	}

	defer rowsUsers.Close()
	for rowsUsers.Next() {
		var id int
		var user UserTopic
		var scanErr = rowsUsers.Scan(&id, &user.Username, &user.Avatar, &user.SignText)

		if scanErr != nil {
			system.ErrLog(errorFunction, scanErr)
			continue
		}

		usersInfo[id] = user
	}

	return usersInfo
}
