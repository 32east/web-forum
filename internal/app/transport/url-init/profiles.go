package url_init

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/middleware"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
	"web-forum/internal/app/templates"
	"web-forum/pkg/stuff"
)

var ctx = context.Background()

func Profiles() {
	middleware.Mult("/profile/([0-9]+)", func(w http.ResponseWriter, r *http.Request, accountId int) {
		var acc, accErr = account.GetById(accountId)

		if accErr != nil {
			middleware.Push404(w, r)
			return
		}

		var infoToSend = r.Context().Value("InfoToSend").(map[string]interface{})
		infoToSend["Title"] = acc.Username

		var timeWithUs = int(math.Round(time.Now().Sub(acc.CreatedAt).Hours() / 24.0))
		var suffix = "дней"

		if timeWithUs == 1.0 {
			suffix = "день"
		} else if timeWithUs >= 2 && timeWithUs <= 4 {
			suffix = "дня"
		}

		var contentToAdd = map[string]interface{}{
			"ProfileUsername":  acc.Username,
			"ProfileCreatedAt": fmt.Sprintf("%d %s", timeWithUs, suffix),
			"ProfileMessages":  []models.ProfileMessage{},
		}

		var sex = "Не указан"

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

		var rowsMessages, errMessages = db.Postgres.Query(ctx, ` 
		select m.topic_id, t.topic_name, m.message, m.create_time
		from (
			select topic_id, message, create_time
		      from messages
		      where account_id = $1
		      order by create_time desc
		      limit 10
		) as m
		inner join topics as t on m.topic_id = t.id
		order by m.create_time desc;
		`, acc.Id)

		if errMessages != nil {
			stuff.ErrLog("initialize_functions.Profiles", fmt.Errorf("Failed to fetch messages: %v", errMessages))
			contentToAdd["ProfileNoMessages"] = true
			templates.ContentAdd(r, templates.Profile, contentToAdd)
			return
		}

		for rowsMessages.Next() {
			msg := models.ProfileMessage{}
			createTime := time.Time{}
			scanErr := rowsMessages.Scan(&msg.TopicId, &msg.TopicName, &msg.Message, &createTime)

			if scanErr != nil {
				stuff.ErrLog("initialize_functions.Profiles", fmt.Errorf("Failed to load message: %v", scanErr))
			}

			msg.CreateTime = createTime.Format("2006-01-02 15:04:05")

			contentToAdd["ProfileMessages"] = append(contentToAdd["ProfileMessages"].([]models.ProfileMessage), msg)
		}

		if len(contentToAdd["ProfileMessages"].([]models.ProfileMessage)) == 0 {
			contentToAdd["ProfileNoMessages"] = true
		}

		templates.ContentAdd(r, templates.Profile, contentToAdd)
	})
}
