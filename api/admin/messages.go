package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"web-forum/system/db"
)

func HandleMessageDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) {
	var receiveId map[string]int
	newDecoder := json.NewDecoder(r.Body).Decode(&receiveId)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, newDecoder.Error()
		return
	}

	val, ok := receiveId["id"]

	if !ok {
		answer["success"], answer["reason"] = false, "id not found"
		return
	}

	tx, err := db.Postgres.Begin(ctx)
	defer func() {
		if answer["success"] == true {
			tx.Commit(ctx)
		} else {
			tx.Rollback(ctx)
		}
	}()

	if err != nil {
		answer["success"], answer["reason"] = false, err.Error()
		return
	}

	var isParentMessage bool
	queryCheck := tx.QueryRow(ctx, "select true from topics where parent_id = $1;", val)
	queryCheck.Scan(&isParentMessage)

	if isParentMessage {
		answer["success"], answer["reason"] = false, "parented message"
		return
	}

	_, execErr := tx.Exec(ctx, "delete from messages where id=$1", val)

	if execErr != nil {
		answer["success"], answer["reason"] = false, execErr.Error()
		return
	}

	fmt.Println(execErr)
	answer["success"] = true
}
