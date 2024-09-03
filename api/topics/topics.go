package topics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"unicode/utf8"
	"web-forum/api/auth"
	"web-forum/internal"
	"web-forum/system/sqlDb"
	"web-forum/www/handlers"
	"web-forum/www/services/account"
)

func HandleMessage(writer *http.ResponseWriter, reader *http.Request) {
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

	accInfo, tokenErr := account.ReadAccountFromCookie(cookie)

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

	_, rowErr := sqlDb.MySqlDB.Query("SELECT id FROM topics WHERE id = ?", topicId)

	if rowErr != nil {
		answer["success"], answer["reason"] = false, "topics-messages not found"

		return
	}

	accountId := accInfo.Id
	_, queryErr := sqlDb.MySqlDB.Exec("INSERT INTO `messages` (id, topic_id, account_id, message, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, NULL)", topicId, accountId, message, time.Now())

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	go func() {
		sqlDb.MySqlDB.Exec("UPDATE `topics` SET message_count = message_count + 1 WHERE id = ?", topicId)
	}()

	answer["success"] = true
}

func HandleTopicCreate(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)
	defer func() {
		if !answer["success"].(bool) {
			log.Println(string(reader.RemoteAddr) + " > on topics-messages create: " + answer["reason"].(string))
		}
	}()

	defer newJSONEncoder.Encode(answer)

	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"

		return
	}

	accInfo, tokenErr := account.ReadAccountFromCookie(cookie)

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
	_, rowErr := sqlDb.MySqlDB.Query("SELECT id FROM forums WHERE id = ?;", categoryId)

	if rowErr != nil {
		answer["success"], answer["reason"] = false, "category not found"

		return
	}

	convertToByte := []byte(name)
	utf8count := 0

	for len(convertToByte) > 0 {
		_, size := utf8.DecodeRune(convertToByte)
		utf8count += 1

		if utf8count >= 128 {
			answer["success"], answer["reason"] = false, fmt.Errorf("max limit of topics-messages name is 128")

			return
		}

		convertToByte = convertToByte[size:]
	}

	currentTime := time.Now()
	rows, queryErr := sqlDb.MySqlDB.Exec("INSERT INTO `topics` (id, forum_id, topic_name, created_by, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, NULL)", categoryId, name, accountId, currentTime)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	lastInsertId, err := rows.LastInsertId()

	if err != nil {
		answer["success"], answer["reason"] = false, "last insert id error"
		return
	}

	_, queryErr = sqlDb.MySqlDB.Exec("INSERT INTO `messages` (id, topic_id, account_id, message, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, NULL)", lastInsertId, accountId, msg, currentTime)

	if queryErr != nil {
		answer["success"], answer["reason"] = false, queryErr.Error()

		return
	}

	newTopicObject := internal.Topic{Id: int(lastInsertId), Name: name, Creator: accountId, CreateTime: currentTime}
	redirect := handlers.CreateTopic(newTopicObject)

	answer["success"], answer["redirect"] = true, redirect
}
