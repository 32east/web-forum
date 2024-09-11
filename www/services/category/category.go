package category

import (
	"context"
	"fmt"
	"log"
	"web-forum/internal"
	"web-forum/system/db"
)

var Cache *[]internal.Category

func Get() (*[]internal.Category, error) {
	if Cache != nil {
		return Cache, nil
	}

	const errorFunction = "category.Get"

	var forums []internal.Category
	rows, err := db.Postgres.Query(context.Background(), "SELECT * FROM forums;")
	defer rows.Close()

	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, err))
	}

	for rows.Next() {
		var id int
		var forumName string
		var forumDescription string

		scanErr := rows.Scan(&id, &forumName, &forumDescription)

		if scanErr != nil {
			log.Fatal(fmt.Errorf("%s [2]: %w", errorFunction, scanErr))
		}

		forums = append(forums, internal.Category{Id: id, Name: forumName, Description: forumDescription})
	}

	Cache = &forums

	return &forums, nil
}
