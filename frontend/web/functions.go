package web

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"html/template"
	"log"
	"math"
	"net/http"
	"reflect"
	"time"
	"web-forum/internal"
)

func ContentAdd(infoToSend *map[string]interface{}, tmpl *template.Template, content any) {
	if reflect.ValueOf(content).Kind() == reflect.Map {
		for k, v := range *infoToSend {
			content.(map[string]interface{})[k] = v
		}
	}

	newBytesBuffer := new(bytes.Buffer)
	tmpl.Execute(newBytesBuffer, content)
	(*infoToSend)["Content"] = template.HTML(newBytesBuffer.String())
}

func TokensRefreshInRedis(reader *http.Request, writer *http.ResponseWriter, rdb *redis.Client) {
	if reader.Referer() != "" {
		return
	}

	ctx := context.Background()
	accessToken, accessCookieErr := reader.Cookie("access_token")

	if accessCookieErr != nil {
		return
	}

	refreshToken, errRefresh := reader.Cookie("refresh_token")

	if errRefresh != nil {
		return
	}

	resultAccessToken, errAccessToken := rdb.Get(ctx, "AToken:"+accessToken.Value).Result()
	resultRefreshToken, errRefreshToken := rdb.Get(ctx, "RToken:"+refreshToken.Value).Result()

	if errAccessToken == nil {
		rdb.Set(ctx, "AToken:"+accessToken.Value, resultAccessToken, time.Hour*12)
	}

	if errRefreshToken == nil {
		rdb.Set(ctx, "RToken:"+refreshToken.Value, resultRefreshToken, time.Hour*72)
	}

	http.SetCookie(*writer, &http.Cookie{
		Name:    "access_token",
		Value:   accessToken.Value,
		Expires: time.Now().Add(time.Hour * 12),
		Path:    "/",
	})

	http.SetCookie(*writer, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken.Value,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 72),
	})
}

func GetForums(db *sql.DB) (*[]interface{}, error) {
	var forums []interface{}
	rows, err := db.Query("SELECT * FROM `forums`;")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int
		var forumName string
		var forumDescription string

		scanErr := rows.Scan(&id, &forumName, &forumDescription)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		forums = append(forums, map[string]interface{}{
			"forum_id":          id,
			"forum_name":        forumName,
			"forum_description": forumDescription,
		})
	}

	return &forums, nil
}

func GetTopics(forumId int, db *sql.DB, page int) (*[]interface{}, int, error) {
	var topics []interface{}

	tx, err := db.Begin()
	defer tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	var topicsCount float64
	queryRow := tx.QueryRow("SELECT COUNT(*) FROM `topics` WHERE forum_id=?;", forumId)
	countTopicsErr := queryRow.Scan(&topicsCount)
	topicsCount = math.Ceil(topicsCount / internal.MaxPaginatorTopics)

	if countTopicsErr != nil {
		log.Fatal(countTopicsErr)
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM `topics` WHERE forum_id = ? ORDER BY id LIMIT %d OFFSET %d;", internal.MaxPaginatorTopics, (page-1)*internal.MaxPaginatorTopics)
	rows, err := tx.Query(fmtQuery, forumId)

	if err != nil {
		log.Fatal(err)
		return &topics, 0, fmt.Errorf("error getting topics: %v", err)
	}

	for rows.Next() {
		var id int
		var forumId int
		var topicName string
		var topicMessage string
		var topicCreator int
		var createTime time.Time
		var updateTime interface{}

		scanErr := rows.Scan(&id, &forumId, &topicName, &topicMessage, &topicCreator, &createTime, &updateTime)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		if updateTime != nil && updateTime.(sql.NullTime).Valid {
			updateTime = updateTime.(time.Time).Format("2006-01-02 15:04:05")
		}

		creatorAccount, ok := internal.GetAccountById(topicCreator)

		if ok != nil {
			log.Fatal("фатальная ошибка при получении креатор аккаунта", topicCreator)
		}

		aboutTopic := map[string]interface{}{
			"topic_id":    id,
			"forum_id":    forumId,
			"topic_name":  topicName,
			"username":    creatorAccount.Username,
			"create_time": createTime.Format("2006-01-02 15:04:05"),
			"update_time": updateTime,
		}

		if creatorAccount.Avatar.Valid {
			aboutTopic["avatar"] = creatorAccount.Avatar.String
		}

		topics = append(topics, aboutTopic)
	}

	return &topics, int(topicsCount), nil
}

func Test_AddTopic(db *sql.DB, forumId int, topicName string, topicMessage string, topicCreator int) {
	_, err := db.Query("INSERT INTO `topics` (id, forum_id, topic_name, topic_message, created_by, create_time, update_time) VALUES (NULL, ?, ?, ?, ?, ?, NULL)", forumId, topicName, topicMessage, topicCreator, time.Now())

	if err != nil {
		log.Fatal(err)
	}
}
