package admin

import (
	"encoding/json"
	"net/http"
	"web-forum/internal"
	"web-forum/system/db"
)

func HandleTopicDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var receiveId = internal.DeleteObject{}
	var newDecoderErr = json.NewDecoder(r.Body).Decode(&receiveId)

	if newDecoderErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return newDecoderErr
	}

	var tx, txErr = db.Postgres.Begin(ctx)

	if txErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return txErr
	}

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	var _, execErr = tx.Exec(ctx, "delete from topics where id=$1", receiveId.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	_, execErr = tx.Exec(ctx, "update categorys set topics_count = topics_count - 1 where id = $1;", receiveId.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	answer["success"] = true

	return nil
}
