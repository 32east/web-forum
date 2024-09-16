package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/www/services/account"
	"web-forum/www/services/category"
)

var ctx = context.Background()

func HandleCategoryCreate(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) {
	cookie, err := r.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return
	}

	_, errGetAccount := account.ReadFromCookie(cookie)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return
	}

	//if !accountData.IsAdmin {
	//	answer["success"], answer["reason"] = false, "no access"
	//	return
	//}

	objCategory := internal.Category{}
	newDecoder := json.NewDecoder(r.Body).Decode(&objCategory)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, newDecoder.Error()
		return
	}

	if objCategory.Name == "" {
		answer["success"], answer["reason"] = false, "objCategory name is empty"
		return
	}

	if objCategory.Description == "" {
		answer["success"], answer["reason"] = false, "objCategory description is empty"
		return
	}

	_, execErr := db.Postgres.Exec(ctx, `insert into forums (forum_name, forum_description) values ($1, $2);`, objCategory.Name, objCategory.Description)

	if execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	category.Cache = nil

	answer["success"] = true
}
