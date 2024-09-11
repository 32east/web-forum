package topics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"web-forum/api/auth"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/www/handlers"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func HandleMessage(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)

	const errFunction = "HandleMessage"
	defer func() {
		if !answer["success"].(bool) {
			system.ErrLog(errFunction, string(reader.RemoteAddr)+" > "+answer["reason"].(string))
		}
	}()

	defer newJSONEncoder.Encode(answer)

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

	row := db.Postgres.QueryRow(ctx, "SELECT id FROM topics WHERE id = $1;", topicId).Scan(&scanTopicId)

	if row != nil {
		answer["success"], answer["reason"] = false, "topic not founded"
		return
	}

	msgInsert := internal.FormatString(message.(string))

	accountId := accInfo.Id
	tx, err := db.Postgres.Begin(ctx)
	defer func() {
		if !answer["success"].(bool) {
			tx.Rollback(ctx)
		}
	}()

	if err != nil {
		answer["success"], answer["reason"] = false, err.Error()
		return
	}

	_, queryErr := tx.Exec(ctx, `insert into messages(topic_id, account_id, message, create_time, update_time) values ($1, $2, $3, current_timestamp, NULL)`, topicId, accountId, msgInsert)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	_, queryErr = tx.Exec(ctx, "update topics set message_count = message_count + 1 where id = $1;", topicId)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	queryErr = tx.Commit(ctx)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	answer["success"] = true
}

func HandleTopicCreate(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)

	const errFunction = "HandleTopicCreate"
	defer func() {
		if !answer["success"].(bool) {
			system.ErrLog(errFunction, string(reader.RemoteAddr)+" > "+answer["reason"].(string))
		}
	}()

	defer newJSONEncoder.Encode(answer)

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
	scanCategoryId := -1

	row := db.Postgres.QueryRow(context.Background(), "SELECT id FROM forums WHERE id = $1;", categoryId).Scan(&scanCategoryId)

	if row != nil {
		answer["success"], answer["reason"] = false, "category not founded"
		return
	}

	if internal.Utf8Length(name) > 128 {
		answer["success"], answer["reason"] = false, "name too long"

		return
	}

	queryErr := db.Postgres.QueryRow(context.Background(), "INSERT INTO topics (forum_id, topic_name, created_by, create_time, update_time) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, NULL) returning id;", categoryId, name, accountId)

	lastInsertId := -1
	errScan := queryErr.Scan(&lastInsertId)

	if errScan != nil {
		answer["success"], answer["reason"] = false, errScan.Error()

		return
	}

	msg = internal.FormatString(msg)
	_, err2Scan := db.Postgres.Exec(context.Background(), "INSERT INTO messages (topic_id, account_id, message, create_time, update_time) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, NULL)", lastInsertId, accountId, msg)

	if err2Scan != nil {
		answer["success"], answer["reason"] = false, err2Scan.Error()

		return
	}

	newTopicObject := internal.Topic{Id: lastInsertId, Name: name, Creator: accountId, CreateTime: time.Now()}
	redirect := handlers.CreateTopic(newTopicObject)

	answer["success"], answer["redirect"] = true, redirect
}
