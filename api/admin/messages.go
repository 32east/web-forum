package admin

import (
	"encoding/json"
	"net/http"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/system/rdb"
)

func HandleMessageDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var receiveId = internal.MessageDelete{}
	var newDecoder = json.NewDecoder(r.Body).Decode(&receiveId)

	if newDecoder != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return newDecoder
	}

	var tx, err = db.Postgres.Begin(ctx)

	if err != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return err
	}

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	var isParentMessage bool
	var queryCheck = tx.QueryRow(ctx, "select true from topics where parent_id = $1;", receiveId.Id)
	queryCheck.Scan(&isParentMessage)

	if isParentMessage {
		answer["success"], answer["reason"] = false, "parented message"
		return nil
	}

	var _, execErr = tx.Exec(ctx, "delete from messages where id=$1", receiveId.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	_, execErr = tx.Exec(ctx, "update topics set message_count = message_count - 1 where id = $1;", receiveId.Id)

	if execErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return execErr
	}

	rdb.RedisDB.Do(ctx, "incrby", "count:messages", -1)
	answer["success"] = true

	return nil
}
