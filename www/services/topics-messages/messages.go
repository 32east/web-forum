package topics_messages

import (
	"context"
	"fmt"
	"log"
	"math"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/www/services/account"
)

var ctx = context.Background()

//func PaginatorQuery(tableName string, columnName string, id int, page int) (pgx.Tx, pgx.Rows, error) {
//	const errorFunction = "PaginatorQuery"
//
//	tx, err := db.Postgres.Begin(ctx)
//	defer tx.Commit(ctx)
//
//	if err != nil {
//		return nil, nil, err
//	}
//
//	var topicsCount float64
//	fmtQuery := fmt.Sprintf("select count(*) from %s where %s = $1", tableName, columnName)
//	queryRow := tx.QueryRow(ctx, fmtQuery, id)
//	countMessagesErr := queryRow.Scan(&topicsCount)
//	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)
//
//	if countMessagesErr != nil {
//		log.Fatal(fmt.Errorf("%s: %w", errorFunction, countMessagesErr))
//	}
//
//	fmtQuery = fmt.Sprintf("SELECT * FROM %s where %s=$1 LIMIT %d OFFSET %d;", tableName, columnName, internal.MaxPaginatorMessages, (page-1)*internal.MaxPaginatorMessages)
//	rows, err := tx.Query(ctx, fmtQuery, id)
//
//	if err != nil {
//		return nil, nil, err
//	}
//
//	paginatorMessages := internal.Paginator{
//		CurrentPage: page,
//		AllPages:    int(pagesCount),
//	}
//
//	return tx, rows, nil
//}

func Get(topic internal.Topic, page int) (*internal.Paginator, error) {
	const errorFunction = "topics_messages.Get"

	tx, err := db.Postgres.Begin(ctx)
	defer tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	var topicsCount float64
	queryRow := tx.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE topic_id = $1;", topic.Id)
	countMessagesErr := queryRow.Scan(&topicsCount)
	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if countMessagesErr != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, countMessagesErr))
	}

	fmtQuery := fmt.Sprintf("SELECT * FROM messages where topic_id=$1 LIMIT %d OFFSET %d;", internal.MaxPaginatorMessages, (page-1)*internal.MaxPaginatorMessages)
	rows, err := tx.Query(ctx, fmtQuery, topic.Id)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	paginatorMessages := internal.Paginator{
		CurrentPage: page,
		AllPages:    int(pagesCount),
	}

	var tempUsers []int
	var tempMessages []internal.Message

	for rows.Next() {
		var msg internal.Message

		scanErr := rows.Scan(&msg.Id, &msg.TopicId, &msg.CreatorId, &msg.Message, &msg.CreateTime, &msg.UpdateTime)

		if scanErr != nil {
			system.ErrLog(errorFunction, scanErr.Error())
			continue
		}

		tempUsers = append(tempUsers, msg.CreatorId)
		tempMessages = append(tempMessages, msg)
	}

	usersInfo := account.GetFromSlice(tempUsers, tx)

	for i := 0; i < len(tempMessages); i++ {
		msg := tempMessages[i]

		acc, ok := usersInfo[msg.CreatorId]

		if !ok {
			system.ErrLog(errorFunction, fmt.Sprintf("Не найден креатор сообщения в бд? > %s(ID): %s(MSG)", msg.CreatorId, msg.TopicId))
			continue
		}

		messageInfo := map[string]interface{}{
			"uid":         msg.CreatorId,
			"username":    acc.Username,
			"message":     msg.Message,
			"create_time": msg.CreateTime.Format("2006-01-02 15:04:05"),
		}

		if msg.UpdateTime.Valid {
			messageInfo["UpdateTime"] = msg.UpdateTime.Time.Format("2006-01-02 15:04:05")
		}

		if acc.Avatar.Valid {
			messageInfo["avatar"] = acc.Avatar.String
		}

		if acc.SignText.Valid {
			messageInfo["sign_text"] = acc.SignText.String
		}

		paginatorMessages.Objects = append(paginatorMessages.Objects, messageInfo)
	}

	return &paginatorMessages, nil
}
