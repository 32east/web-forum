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

func HandleCategoryCreate(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) {
	objCategory := internal.Category{}
	newDecoder := json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, newDecoder.Error()
		return
	case objCategory.Name == "":
		answer["success"], answer["reason"] = false, "objCategory name is empty"
		return
	case objCategory.Description == "":
		answer["success"], answer["reason"] = false, "objCategory description is empty"
		return
	default:
	}

	_, execErr := db.Postgres.Exec(ctx, `insert into forums (forum_name, forum_description) values ($1, $2);`, objCategory.Name, objCategory.Description)

	if execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	category.Cache = nil

	answer["success"] = true
}

func HandleCategoryEdit(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) {
	objCategory := internal.Category{}
	newDecoder := json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, newDecoder.Error()
		return
	case objCategory.Id <= 0:
		answer["success"], answer["reason"] = false, "category not found"
		return
	case objCategory.Name == "":
		answer["success"], answer["reason"] = false, "objCategory name is empty"
		return
	case objCategory.Description == "":
		answer["success"], answer["reason"] = false, "objCategory description is empty"
		return
	default:
	}

	_, execErr := db.Postgres.Exec(ctx, `update forums set forum_name = $1, forum_description = $2 where id = $3;`, objCategory.Name, objCategory.Description, objCategory.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	category.Cache = nil

	answer["success"] = true
}

func HandleCategoryDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) {
	var query map[string]int
	newDecoder := json.NewDecoder(r.Body).Decode(&query)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, newDecoder.Error()
		return
	}

	_, execErr := db.Postgres.Exec(ctx, `delete from forums where id = $1;`, query["id"])

	if execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	category.Cache = nil

	answer["success"] = true
}
