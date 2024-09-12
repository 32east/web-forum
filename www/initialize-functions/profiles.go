package initialize_functions

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/www/handlers"
	"web-forum/www/services/account"
	"web-forum/www/templates"
)

var ctx = context.Background()

func CreateProfilePage(accountId int) {
	http.HandleFunc("/profile/"+fmt.Sprint(accountId)+"/", func(w http.ResponseWriter, r *http.Request) {
		infoToSend, _ := handlers.Base(r, &w)
		acc, accErr := account.GetById(accountId)
		if accErr != nil {
			system.ErrLog("initialize_functions.Profiles", fmt.Sprintf("Failed to load profile: %v", accErr))
		}

		(*infoToSend)["Title"] = acc.Username
		defer templates.Index.Execute(w, infoToSend)

		timeWithUs := int(math.Round(time.Now().Sub(acc.CreatedAt).Hours() / 24.0))
		suffix := "дней"

		if timeWithUs == 1.0 {
			suffix = "день"
		} else if timeWithUs >= 2 && timeWithUs <= 4 {
			suffix = "дня"
		}
		contentToAdd := map[string]interface{}{
			"ProfileUsername":  acc.Username,
			"ProfileCreatedAt": fmt.Sprintf("%d %s", timeWithUs, suffix),
			"ProfileMessages":  []internal.ProfileMessage{},
		}

		sex := "Не указан"

		if acc.Sex.String == "m" {
			sex = "Мужской"
		} else if acc.Sex.String == "f" {
			sex = "Женский"
		}

		contentToAdd["ProfileSex"] = sex

		if acc.SignText.Valid {
			contentToAdd["ProfileSignText"] = acc.SignText.String
		}

		if acc.Avatar.Valid {
			contentToAdd["ProfileAvatar"] = acc.Avatar.String
		}

		if acc.Description.Valid {
			contentToAdd["ProfileDescription"] = acc.Description.String
		}

		rowsMessages, errMessages := db.Postgres.Query(ctx, `
				select m.topic_id, t.topic_name, m.message, m.create_time
				from messages as m
				inner join topics as t on m.topic_id = t.id
				where account_id = $1
				order by m.create_time desc
				limit 10;`, acc.Id)

		if errMessages != nil {
			system.ErrLog("initialize_functions.Profiles", fmt.Sprintf("Failed to fetch messages: %v", errMessages))
		}

		for rowsMessages.Next() {
			msg := internal.ProfileMessage{}
			createTime := time.Time{}
			scanErr := rowsMessages.Scan(&msg.TopicId, &msg.TopicName, &msg.Message, &createTime)

			if scanErr != nil {
				system.ErrLog("initialize_functions.Profiles", fmt.Sprintf("Failed to load message: %v", scanErr))
			}

			msg.CreateTime = createTime.Format("2006-01-02 15:04:05")

			contentToAdd["ProfileMessages"] = append(contentToAdd["ProfileMessages"].([]internal.ProfileMessage), msg)
		}

		if len(contentToAdd["ProfileMessages"].([]internal.ProfileMessage)) == 0 {
			contentToAdd["ProfileNoMessages"] = true
		}

		templates.ContentAdd(infoToSend, templates.Profile, contentToAdd)
	})
}

func Profiles() {
	rows, err := db.Postgres.Query(ctx, "select * from users;")

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		accountId := -1

		err = rows.Scan(&accountId, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		if err != nil {
			system.ErrLog("initialize_functions.Profiles", fmt.Sprintf("Failed to initialize profile: %v", err))
			continue
		}

		fmt.Println(accountId)
		CreateProfilePage(accountId)
	}
}
