package initialize_functions

import (
	"log"
	"web-forum/internal"
	"web-forum/system/sqlDb"
	"web-forum/www/handlers"
)

func InitializeTopicsPages() {
	rows, err := sqlDb.MySqlDB.Query("SELECT * FROM `topics`;")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		topic := internal.Topic{}
		scanErr := rows.Scan(&topic.Id, &topic.ForumId, &topic.Name, &topic.Creator, &topic.CreateTime, &topic.UpdateTime, &topic.MessageCount)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		handlers.CreateTopic(topic)
	}
}
