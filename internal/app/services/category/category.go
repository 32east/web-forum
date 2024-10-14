package category

import (
	"context"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/models"
	"web-forum/pkg/stuff"
)

var Cache *[]models.Category
var CacheMap = make(map[int]*models.Category)
var ctx = context.Background()

func GetAll() (*[]models.Category, error) {
	if Cache != nil {
		return Cache, nil
	}

	const errorFunction = "category.GetAll"

	var categorys []models.Category
	var rows, err = db.Postgres.Query(ctx, "select * from categorys order by id;")

	if err != nil {
		stuff.FatalLog(errorFunction, err)
	}

	defer rows.Close()

	for rows.Next() {
		var category = models.Category{}
		var scanErr = rows.Scan(&category.Id, &category.Name, &category.Description, &category.TopicsCount)

		if scanErr != nil {
			stuff.FatalLog(errorFunction, scanErr)
		}

		categorys = append(categorys, category)
	}

	Cache = &categorys

	for _, value := range *Cache {
		CacheMap[value.Id] = &value
	}

	return &categorys, nil
}

func GetInfo(id int) *models.Category {
	var val, ok = CacheMap[id]

	if !ok {
		return &models.Category{}
	}

	return val
}
