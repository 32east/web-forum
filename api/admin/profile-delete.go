package admin

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
	"os"
	"strconv"
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/www/services/account"
)

type objectToRemove struct {
	// Да. В БД можно было два одинаковых названия сделать.
	// Я знаю. Мне лень.
	AccountColumn string
	AccountID     int

	ColumnParent string
	TableName    string
	CountColumn  string
	Update       string

	Tx *pgx.Tx
}

func countMinus(object *objectToRemove) error {
	var tx = *object.Tx
	var fmtQuery = fmt.Sprintf(`select %s from %s where %s = $1`, object.ColumnParent, object.TableName, object.AccountColumn)
	var msgRows, msgErr = tx.Query(ctx, fmtQuery, object.AccountID)

	if msgErr != nil {
		return msgErr
	}

	// ObjectId | Count
	var affected = make(map[int]int)

	for msgRows.Next() {
		var objectId int
		var errRow = msgRows.Scan(&objectId)

		if errRow != nil {
			return errRow
		}

		affected[objectId] += 1
	}

	for objectId, minusCount := range affected {
		var fmtExec = fmt.Sprintf(`update %s set %s = %s - $1 where id = $2`, object.Update, object.CountColumn, object.CountColumn)
		var _, execDeleteErr = tx.Exec(ctx, fmtExec, minusCount, objectId)

		if execDeleteErr != nil {
			return execDeleteErr
		}
	}

	return nil
}

func HandleProfileDelete(w http.ResponseWriter, r *http.Request, answer map[string]interface{}) error {
	var id = r.FormValue("id")
	var conv, err = strconv.Atoi(id)

	if err != nil {
		answer["success"], answer["reason"] = false, "invalid id"
		return nil
	}

	var accountData, errGetAccount = account.GetById(conv)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "invalid user"
		return nil
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

	var rmMessages = &objectToRemove{
		AccountColumn: "account_id",
		AccountID:     conv,
		ColumnParent:  "topic_id",
		TableName:     "messages",
		Update:        "topics",
		CountColumn:   "message_count",
		Tx:            &tx,
	}

	var minusMsgErr = countMinus(rmMessages)

	if minusMsgErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return minusMsgErr
	}

	var rmCategory = &objectToRemove{
		AccountID:     conv,
		AccountColumn: "created_by",
		ColumnParent:  "forum_id",
		TableName:     "topics",
		Update:        "categorys",
		CountColumn:   "topics_count",
		Tx:            &tx,
	}

	var minusCatErr = countMinus(rmCategory)

	if minusCatErr != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return minusCatErr
	}

	var cmd, cmdErr = tx.Exec(ctx, `delete from users where id = $1;`, conv)

	if cmdErr != nil {
		answer["success"], answer["reason"] = false, "invalid user"
		return cmdErr
	}

	var affected = cmd.RowsAffected()

	if affected == 0 {
		answer["success"], answer["reason"] = false, "invalid user"
		return nil
	}

	os.Remove(internal.AvatarsFilePath + accountData.Avatar.String)

	answer["success"] = true
	rdb.RedisDB.Del(ctx, fmt.Sprintf("aID:%d", conv))
	delete(account.FastCache, conv)

	return nil
}
