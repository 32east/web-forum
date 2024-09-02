package topics

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
	"web-forum/api/auth"
	"web-forum/api/message"
	"web-forum/internal"
	"web-forum/www/services/account"
)

func HandleMessage(writer *http.ResponseWriter, reader *http.Request, db *sql.DB, rdb *redis.Client) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)
	defer func() {
		if !answer["success"].(bool) {
			log.Println(string(reader.RemoteAddr) + " > on message send: " + answer["reason"].(string))
		}
	}()

	defer newJSONEncoder.Encode(answer)

	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	login, ok := rdb.Get(context.Background(), "AToken:"+cookie.Value).Result()

	if ok != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	accInfo, errGetAccount := account.GetAccount(login)

	if errGetAccount != nil {
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
	_, rowErr := db.Query("SELECT id FROM topics WHERE id = ?", topicId)

	if rowErr != nil {
		answer["success"], answer["reason"] = false, "topic not found"

		return
	}

	accountId := accInfo.Id
	_, queryErr := db.Exec("INSERT INTO `messages` (id, topic_id, account_id, message, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, NULL)", topicId, accountId, message, time.Now())

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	answer["success"] = true
}

func HandleTopicCreate(writer *http.ResponseWriter, reader *http.Request, db *sql.DB, rdb *redis.Client) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)
	defer func() {
		if !answer["success"].(bool) {
			log.Println(string(reader.RemoteAddr) + " > on topic create: " + answer["reason"].(string))
		}
	}()

	defer newJSONEncoder.Encode(answer)

	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	login, ok := rdb.Get(context.Background(), "AToken:"+cookie.Value).Result()

	if ok != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	accInfo, err := account.GetAccount(login)

	if err != nil {
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
	_, rowErr := db.Query("SELECT id FROM forums WHERE id = ?;", categoryId)

	if rowErr != nil {
		answer["success"], answer["reason"] = false, "category not found"

		return
	}

	currentTime := time.Now()
	rows, queryErr := db.Exec("INSERT INTO `topics` (id, forum_id, topic_name, topic_message, created_by, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, ?, NULL)", categoryId, name, msg, accountId, currentTime)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	lastInsertId, err := rows.LastInsertId()

	if err != nil {
		answer["success"], answer["reason"] = false, "last insert id error"
		return
	}

	newTopicObject := internal.Topic{Id: int(lastInsertId), Name: name, Message: msg, Creator: accountId, CreateTime: currentTime}
	redirect := message.CreateTopic(newTopicObject, db, rdb)

	answer["success"], answer["redirect"] = true, redirect
}
