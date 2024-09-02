package web

import (
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
	"web-forum/internal"
)

func CreateTopic(topic internal.Topic, db *sql.DB, rdb *redis.Client) string {
	url := "/topics/" + fmt.Sprint(topic.Id)

	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		infoToSend, _ := HandleBase(r, &w, rdb)
		(*infoToSend)["Title"] = topic.Name // TODO: Ограничить символы до 128
		defer IndexTemplate.Execute(w, infoToSend)

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

			getAccount, ok := internal.GetAccountById(accountId)

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

		getAccount, ok := internal.GetAccountById(topic.Creator)

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

		ContentAdd(infoToSend, TopicTemplate, topicInfo)
	})

	return url
}

func InitializeForumsPages(db *sql.DB, rdb *redis.Client) {
	forums, err := GetForums(db)

	if err != nil {
		panic(err)
	}

	for _, output := range *forums {
		outputToMap := output.(map[string]interface{})
		forumId := outputToMap["forum_id"].(int)

		http.HandleFunc("/category/"+fmt.Sprint(forumId)+"/", func(w http.ResponseWriter, r *http.Request) {
			currentPage := r.FormValue("page")

			infoToSend, _ := HandleBase(r, &w, rdb)
			(*infoToSend)["Title"] = outputToMap["forum_name"].(string)
			defer IndexTemplate.Execute(w, infoToSend)

			if currentPage == "" {
				currentPage = "1"
			}

			currentPageInt, errInt := strconv.Atoi(currentPage)

			if errInt != nil {
				currentPageInt = 0
			}

			topics, ourPages, topicsErr := GetTopics(forumId, db, currentPageInt)

			if topicsErr != nil {
				log.Fatal(topicsErr)
			}

			categoryIsEmpty := len(*topics) == 0
			howMuchPagesWillBeVisible := internal.HowMuchPagesWillBeVisibleInPaginator
			dividedBy2 := float64(howMuchPagesWillBeVisible) / 2
			floorDivided := int(math.Floor(dividedBy2))
			ceilDivided := int(math.Ceil(dividedBy2))

			if ourPages < internal.HowMuchPagesWillBeVisibleInPaginator {
				howMuchPagesWillBeVisible = ourPages
			}

			if currentPageInt > ourPages {
				currentPageInt = ourPages
			}

			currentPageInt = currentPageInt - 1 // Массив с нуля начинается.
			limitMin, limitMax := currentPageInt-floorDivided, currentPageInt+floorDivided

			if limitMin < 0 {
				limitMin = 0
			}

			if limitMax > ourPages-1 {
				limitMax = ourPages - 1
			}

			if currentPageInt < ceilDivided {
				limitMax = howMuchPagesWillBeVisible - 1
			} else if currentPageInt >= ourPages-ceilDivided {
				limitMin = ourPages - howMuchPagesWillBeVisible
			}

			paginatorPages := make([]int, limitMax-limitMin+1)
			paginatorKey := 0

			for showedPage := limitMin; showedPage <= limitMax; showedPage++ {
				paginatorPages[paginatorKey] = showedPage + 1
				paginatorKey += 1
			}

			contentToSend := map[string]interface{}{
				"forum_id":          forumId,
				"forum_name":        outputToMap["forum_name"],
				"topics":            topics,
				"category_is_empty": categoryIsEmpty,
				"current_page":      currentPageInt,
				"paginator":         paginatorPages,
			}

			currentPageInt += 1

			if currentPageInt > 1 {
				contentToSend["paginator_left"] = currentPageInt - 1
			}

			if currentPageInt < ourPages {
				contentToSend["paginator_right"] = currentPageInt + 1
			}

			ContentAdd(infoToSend, TopicsTemplate, contentToSend)
		})
	}
}

func InitializeTopicsPages(db *sql.DB, rdb *redis.Client) {
	rows, err := db.Query("SELECT * FROM `topics`;")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		topic := internal.Topic{}
		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Message, &topic.Creator, &topic.CreateTime, &topic.UpdateTime)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		CreateTopic(topic, db, rdb)
	}
}
