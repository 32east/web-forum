package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/services/category"
)

var ctx = context.Background()

func HandleCategoryCreate(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	objCategory := internal.Category{}
	newDecoder := json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, "internal server error"
		return newDecoder
	case objCategory.Name == "":
		answer["success"], answer["reason"] = false, "category name is empty"
		return nil
	case objCategory.Description == "":
		answer["success"], answer["reason"] = false, "category description is empty"
		return nil
	default:
	}

	_, execErr := db.Postgres.Exec(ctx, `insert into forums (forum_name, forum_description) values ($1, $2);`, objCategory.Name, objCategory.Description)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true

	return nil
}

func HandleCategoryEdit(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	objCategory := internal.Category{}
	newDecoder := json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, "internal server error"
		return newDecoder
	case objCategory.Id <= 0:
		answer["success"], answer["reason"] = false, "category not founded"
		return nil
	case objCategory.Name == "":
		answer["success"], answer["reason"] = false, "category name is empty"
		return nil
	case objCategory.Description == "":
		answer["success"], answer["reason"] = false, "category description is empty"
		return nil
	default:
	}

	_, execErr := db.Postgres.Exec(ctx, `update forums set forum_name = $1, forum_description = $2 where id = $3;`, objCategory.Name, objCategory.Description, objCategory.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true
	return nil
}

func HandleCategoryDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	query := internal.CategoryDelete{}
	newDecoder := json.NewDecoder(r.Body).Decode(&query)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return newDecoder
	}

	_, execErr := db.Postgres.Exec(ctx, `delete from forums where id = $1;`, query.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true

	return nil
}
