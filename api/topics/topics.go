package topics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func HandleMessage(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) {
	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	accInfo, tokenErr := account.ReadFromCookie(cookie)

	if tokenErr != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	jsonData := map[string]interface{}{
		"message":  "",
		"topic_id": -1,
	}

	jsonErr := json.NewDecoder(reader.Body).Decode(&jsonData)

	if jsonErr != nil {
		answer["success"], answer["reason"] = false, jsonErr.Error()

		return
	}

	topicId, message := jsonData["topic_id"], jsonData["message"]
	scanTopicId := -1

	tx, err := db.Postgres.Begin(ctx)

	if err != nil {
		answer["success"], answer["reason"] = false, err.Error()
		return
	}

	errScan := tx.QueryRow(ctx, "select id from topics where id = $1;", topicId).Scan(&scanTopicId)

	if errScan != nil {
		answer["success"], answer["reason"] = false, errScan.Error()
		return
	}

	msgInsert := internal.FormatString(message.(string))
	accountId := accInfo.Id

	defer func() {
		if !answer["success"].(bool) {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	_, queryErr := tx.Exec(ctx, `insert into messages(topic_id, account_id, message, create_time, update_time) values ($1, $2, $3, current_timestamp, NULL)`, topicId, accountId, msgInsert)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	messageCount := 0
	queryErr = tx.QueryRow(ctx, `update topics
	set message_count = message_count + 1
	where id = $1
	returning message_count;`, topicId).Scan(&messageCount)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	pagesCount := math.Ceil(float64((messageCount)/internal.MaxPaginatorMessages)) + 1

	if pagesCount > 1 {
		answer["page"] = int(pagesCount)
	}

	answer["success"] = true
}

func HandleTopicCreate(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) {
	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	accInfo, tokenErr := account.ReadFromCookie(cookie)

	if tokenErr != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	topic := map[string]string{}
	jsonErr := json.NewDecoder(reader.Body).Decode(&topic)

	if jsonErr != nil {
		answer["success"], answer["reason"] = false, jsonErr.Error()

		return
	}

	name, msg, categoryId, accountId := topic["name"], topic["message"], topic["category_id"], accInfo.Id

	if internal.Utf8Length(name) > 128 {
		answer["success"], answer["reason"] = false, "name too long"

		return
	}

	scanCategoryId := -1

	tx, err := db.Postgres.Begin(ctx)
	defer func() {
		if !answer["success"].(bool) {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	row := tx.QueryRow(ctx, "select id from forums where id = $1;", categoryId).Scan(&scanCategoryId)

	if row != nil {
		answer["success"], answer["reason"] = false, "category not founded"
		return
	}

	queryErr := tx.QueryRow(ctx, "insert into topics (forum_id, topic_name, created_by, create_time, parent_id) values ($1, $2, $3, now(), -1) returning id;", categoryId, name, accountId)

	lastInsertId := -1
	errScan := queryErr.Scan(&lastInsertId)

	if errScan != nil {
		answer["success"], answer["reason"] = false, errScan.Error()

		return
	}

	msg = internal.FormatString(msg)
	msgIdQuery := tx.QueryRow(ctx, "insert into messages (topic_id, account_id, message, create_time) values ($1, $2, $3, now()) returning id;", lastInsertId, accountId, msg)

	var msgId int
	msgIdQuery.Scan(&msgId)

	if _, execErr := tx.Exec(ctx, "update topics set parent_id = $1 where id = $2;", msgId, lastInsertId); execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	go func() {
		db.Postgres.Exec(ctx, "update forums set topics_count = topics_count + 1 where id = $1;", categoryId)
	}()

	answer["success"], answer["redirect"] = true, fmt.Sprintf("/topics/%d", lastInsertId)
}
