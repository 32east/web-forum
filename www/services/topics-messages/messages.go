package topics_messages

import (
	"context"
	"fmt"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/www/services/account"
	"web-forum/www/services/paginator"
)

var ctx = context.Background()

func Get(topic internal.Topic, page int) (*internal.Paginator, error) {
	const errorFunction = "topics_messages.Get"

	queryCount := fmt.Sprintf("select message_count from topics where id = %d", topic.Id)
	tx, rows, paginatorMessages, err := paginator.Query("messages",
		"id, topic_id, account_id, message, create_time, update_time",
		"topic_id", topic.Id, page, queryCount)

	fmt.Println(err)
	defer tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	var tempUsers []int
	var tempMessages []internal.Message

	for rows.Next() {
		var msg internal.Message

		scanErr := rows.Scan(&msg.Id, &msg.TopicId, &msg.CreatorId, &msg.Message, &msg.CreateTime, &msg.UpdateTime)

		if scanErr != nil {
			system.ErrLog(errorFunction, scanErr)
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
			system.ErrLog(errorFunction, fmt.Errorf("Не найден креатор сообщения в бд? > %s(ID): %s(MSG)", msg.CreatorId, msg.TopicId))
			continue
		}

		messageInfo := map[string]interface{}{
			"MessageId":  msg.Id,
			"CreatorId":  msg.CreatorId,
			"Username":   acc.Username,
			"Message":    msg.Message,
			"CreateTime": msg.CreateTime.Format("2006-01-02 15:04:05"),
		}

		if msg.UpdateTime.Valid {
			messageInfo["UpdateTime"] = msg.UpdateTime.Time.Format("2006-01-02 15:04:05")
		}

		if acc.Avatar.Valid {
			messageInfo["Avatar"] = acc.Avatar.String
		}

		if acc.SignText.Valid {
			messageInfo["SignText"] = acc.SignText.String
		}

		paginatorMessages.Objects = append(paginatorMessages.Objects, messageInfo)
	}

	return &paginatorMessages, nil
}
