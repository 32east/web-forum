package initialize_functions

import (
	"context"
	"fmt"
	"log"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/handlers"
)

func Topics() {
	const errorFunc = "InitializeTopicsPages"
	rows, err := db.Postgres.Query(context.Background(), "SELECT * FROM topics;")
	defer rows.Close()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunc, err))
	}

	for rows.Next() {
		topic := internal.Topic{}
		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount)

		if scanErr != nil {
			log.Fatal(fmt.Errorf("%s [2]: %w", errorFunc, scanErr))
		}

		handlers.CreateTopic(topic)
	}
}
