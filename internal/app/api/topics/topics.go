package topics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/functions"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
)

var ctx = context.Background()

func HandleMessage(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	var cookie, err = reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	var accInfo, tokenErr = account.ReadFromCookie(cookie)

	if tokenErr != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	var jsonData = models.MessageCreate{}
	var jsonErr = json.NewDecoder(reader.Body).Decode(&jsonData)

	if jsonErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return jsonErr
	}

	var topicId, message = jsonData.TopicId, jsonData.Message

	if strings.Trim(message, " ") == "" {
		answer["success"], answer["reason"] = false, "message is empty"
		return nil
	}

	var scanTopicId = -1
	var tx, txErr = db.Postgres.Begin(ctx)

	if txErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return txErr
	}

	var errScan = tx.QueryRow(ctx, "select id from topics where id = $1;", topicId).Scan(&scanTopicId)

	if errScan != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		fmt.Println(errScan)
		return errScan
	}

	var msgInsert = functions.FormatString(message)
	var accountId = accInfo.Id

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	var _, queryErr = tx.Exec(ctx, `insert into messages(topic_id, account_id, message, create_time, update_time) values ($1, $2, $3, current_timestamp, NULL)`, topicId, accountId, msgInsert)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return queryErr
	}

	var messageCount = 0
	queryErr = tx.QueryRow(ctx, `update topics
	set message_count = message_count + 1
	where id = $1
	returning message_count;`, topicId).Scan(&messageCount)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return queryErr
	}

	var pagesCount = math.Ceil(float64((messageCount)/models.MaxPaginatorMessages)) + 1

	if pagesCount > 1 {
		answer["page"] = int(pagesCount)
	}

	answer["success"] = true

	return nil
}

func HandleTopicCreate(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	var cookie, err = reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	var accInfo, tokenErr = account.ReadFromCookie(cookie)

	if tokenErr != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	var topic = models.TopicCreate{}
	var jsonErr = json.NewDecoder(reader.Body).Decode(&topic)

	if jsonErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return jsonErr
	}

	var name, msg, categoryId, accountId = topic.Name, topic.Message, topic.CategoryId, accInfo.Id

	if strings.Trim(name, " ") == "" {
		answer["success"], answer["reason"] = false, "topic name is empty"
		return nil
	}

	if strings.Trim(msg, " ") == "" {
		answer["success"], answer["reason"] = false, "message is empty"
		return nil
	}

	if functions.Utf8Length(name) > 128 {
		answer["success"], answer["reason"] = false, "name too long"
		return nil
	}

	var scanCategoryId = -1
	var tx, beginErr = db.Postgres.Begin(ctx)

	if beginErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return beginErr
	}

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	var row = tx.QueryRow(ctx, "select id from categorys where id = $1;", categoryId).Scan(&scanCategoryId)

	if row != nil {
		answer["success"], answer["reason"] = false, "category not founded"
		return nil
	}

	var queryErr = tx.QueryRow(ctx, "insert into topics (forum_id, topic_name, created_by, create_time, parent_id) values ($1, $2, $3, now(), NULL) returning id;", categoryId, name, accountId)
	var lastInsertId = -1
	var errScan = queryErr.Scan(&lastInsertId)

	if errScan != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return errScan
	}

	msg = functions.FormatString(msg)
	var msgIdQuery = tx.QueryRow(ctx, "insert into messages (topic_id, account_id, message, create_time) values ($1, $2, $3, now()) returning id;", lastInsertId, accountId, msg)

	var msgId int
	msgIdQuery.Scan(&msgId)

	if _, execErr := tx.Exec(ctx, "update topics set parent_id = $1 where id = $2;", msgId, lastInsertId); execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	go func() {
		db.Postgres.Exec(ctx, "update categorys set topics_count = topics_count + 1 where id = $1;", categoryId)
		rdb.RedisDB.Do(ctx, "incrby", "count:topics", 1)
	}()

	answer["success"], answer["redirect"] = true, fmt.Sprintf("/topics/%d", lastInsertId)

	return nil
}
