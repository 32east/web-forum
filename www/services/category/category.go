package category

import (
	"context"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
)

var Cache *[]internal.Category
var CacheMap = make(map[int]*internal.Category)

func GetAll() (*[]internal.Category, error) {
	if Cache != nil {
		return Cache, nil
	}

	const errorFunction = "category.GetAll"

	var forums []internal.Category
	rows, err := db.Postgres.Query(context.Background(), "SELECT * FROM forums;")
	defer rows.Close()

	if err != nil {
		system.FatalLog(errorFunction, err)
	}

	for rows.Next() {
		category := internal.Category{}
		scanErr := rows.Scan(&category.Id, &category.Name, &category.Description, &category.TopicsCount)

		if scanErr != nil {
			system.FatalLog(errorFunction, scanErr)
		}

		forums = append(forums, category)
	}

	Cache = &forums

	for _, value := range *Cache {
		CacheMap[value.Id] = &value
	}

	return &forums, nil
}

func GetInfo(id int) *internal.Category {
	val, ok := CacheMap[id]

	if !ok {
		return &internal.Category{}
	}

	return val
}
