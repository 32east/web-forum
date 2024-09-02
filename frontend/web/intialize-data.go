package web

import (
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"math"
	"net/http"
	"strconv"
	"web-forum/api/message"
	"web-forum/internal"
	"web-forum/www/handlers"
	"web-forum/www/services/category"
	"web-forum/www/templates"
)

func InitializeForumsPages(db *sql.DB, rdb *redis.Client) {
	forums, err := category.GetForums(db)

	if err != nil {
		panic(err)
	}

	for _, output := range *forums {
		outputToMap := output.(map[string]interface{})
		forumId := outputToMap["forum_id"].(int)

		http.HandleFunc("/category/"+fmt.Sprint(forumId)+"/", func(w http.ResponseWriter, r *http.Request) {
			currentPage := r.FormValue("page")

			infoToSend, _ := handlers.HandleBase(r, &w)
			(*infoToSend)["Title"] = outputToMap["forum_name"].(string)
			defer templates.IndexTemplate.Execute(w, infoToSend)

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

			templates.ContentAdd(infoToSend, templates.TopicsTemplate, contentToSend)
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

		message.CreateTopic(topic, db, rdb)
	}
}
