package category

import (
	"context"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
)

var Cache *[]internal.Category
var CacheMap = make(map[int]*internal.Category)
var ctx = context.Background()

func GetAll() (*[]internal.Category, error) {
	if Cache != nil {
		return Cache, nil
	}

	const errorFunction = "category.GetAll"

	var categorys []internal.Category
	var rows, err = db.Postgres.Query(ctx, "select * from categorys order by id;")

	if err != nil {
		system.FatalLog(errorFunction, err)
	}

	defer rows.Close()

	for rows.Next() {
		var category = internal.Category{}
		var scanErr = rows.Scan(&category.Id, &category.Name, &category.Description, &category.TopicsCount)

		if scanErr != nil {
			system.FatalLog(errorFunction, scanErr)
		}

		categorys = append(categorys, category)
	}

	Cache = &categorys

	for _, value := range *Cache {
		CacheMap[value.Id] = &value
	}

	return &categorys, nil
}

func GetInfo(id int) *internal.Category {
	var val, ok = CacheMap[id]

	if !ok {
		return &internal.Category{}
	}

	return val
}
