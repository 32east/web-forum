package message

import (
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
	"web-forum/www/handlers"
	"web-forum/www/services/account"
	"web-forum/www/templates"

	"web-forum/internal"
)

func CreateTopic(topic internal.Topic, db *sql.DB, rdb *redis.Client) string {
	url := "/topics/" + fmt.Sprint(topic.Id)

	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		infoToSend, _ := handlers.HandleBase(r, &w)
		(*infoToSend)["Title"] = topic.Name // TODO: Ограничить символы до 128
		defer templates.IndexTemplate.Execute(w, infoToSend)

		rows, err := db.Query("SELECT * FROM `messages` where topic_id=?", topic.Id)

		if err != nil {
			log.Fatal(err)
		}

		var topicMessages []map[string]interface{}

		for rows.Next() {
			var id int
			var topicId int
			var accountId int
			var message string
			var createTime time.Time
			var updateTime interface{}

			rows.Scan(&id, &topicId, &accountId, &message, &createTime, &updateTime)

			getAccount, ok := account.GetAccountById(accountId)

			if ok != nil {
				log.Fatal("фатальная ошибка при получении информации об аккаунте:", accountId)
			}

			if updateTime != nil && updateTime.(sql.NullTime).Valid {
				updateTime = updateTime.(time.Time).Format("2006-01-02 15:04:05")
			}

			messageInfo := map[string]interface{}{
				"username":    getAccount.Username,
				"message":     message,
				"create_time": createTime.Format("2006-01-02 15:04:05"),
				"update_time": updateTime,
			}

			if getAccount.Avatar.Valid {
				messageInfo["avatar"] = getAccount.Avatar.String
			}

			if getAccount.SignText.Valid {
				messageInfo["sign_text"] = getAccount.SignText.String
			}

			topicMessages = append(topicMessages, messageInfo)
		}

		getAccount, ok := account.GetAccountById(topic.Creator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении информации о создателе топика", topic.Creator)
		}

		topicInfo := map[string]interface{}{
			"topic_name":  topic.Name,
			"message":     topic.Message,
			"username":    getAccount.Username,
			"create_time": topic.CreateTime.Format("2006-01-02 15:04:05"),
			"messages":    topicMessages,
		}

		if getAccount.Avatar.Valid {
			topicInfo["avatar"] = getAccount.Avatar.String
		}

		if getAccount.SignText.Valid {
			topicInfo["sign_text"] = getAccount.SignText.String
		}

		templates.ContentAdd(infoToSend, templates.TopicTemplate, topicInfo)
	})

	return url
}
