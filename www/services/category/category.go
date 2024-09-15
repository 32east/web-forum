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
		category := internal.Category{}
		scanErr := rows.Scan(&category.Id, &category.Name, &category.Description, &category.TopicsCount)

		if scanErr != nil {
			log.Fatal(fmt.Errorf("%s [2]: %w", errorFunction, scanErr))
		}

		forums = append(forums, category)
	}

	Cache = &forums

	return &forums, nil
}
