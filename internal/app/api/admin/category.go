package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/category"
)

var ctx = context.Background()

func HandleCategoryCreate(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var objCategory = models.Category{}
	var newDecoder = json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return newDecoder
	case objCategory.Name == "":
		answer["success"], answer["reason"] = false, "category name is empty"
		return nil
	case objCategory.Description == "":
		answer["success"], answer["reason"] = false, "category description is empty"
		return nil
	default:
	}

	var _, execErr = db.Postgres.Exec(ctx, `insert into categorys (name, description) values ($1, $2);`, objCategory.Name, objCategory.Description)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true

	return nil
}

func HandleCategoryEdit(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var objCategory = models.Category{}
	var newDecoder = json.NewDecoder(r.Body).Decode(&objCategory)

	switch {
	case newDecoder != nil:
		answer["success"], answer["reason"] = false, "const-funcs server error"
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

	var _, execErr = db.Postgres.Exec(ctx, `update categorys set name = $1, description = $2 where id = $3;`, objCategory.Name, objCategory.Description, objCategory.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true
	return nil
}

func HandleCategoryDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var query = models.DeleteObject{}
	var newDecoder = json.NewDecoder(r.Body).Decode(&query)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return newDecoder
	}

	var _, execErr = db.Postgres.Exec(ctx, `delete from categorys where id = $1;`, query.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "const-funcs server error"
		return execErr
	}

	category.Cache = nil
	answer["success"] = true

	return nil
}
