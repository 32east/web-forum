package account

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const SeparateChar = ">!;"

func (data Account) Serialize() string {
	var avatar string
	var description string
	var signText string
	var createdAt string
	var updatedAt string

	if data.Avatar.Valid {
		avatar = data.Avatar.String
	}

	if data.SignText.Valid {
		signText = data.SignText.String
	}

	if data.Description.Valid {
		description = data.Description.String
	}

	createdAt = data.CreatedAt.Format("2006-01-02 15:04:05")

	if data.UpdatedAt.Valid {
		updatedAt = data.UpdatedAt.Time.Format(time.RFC3339)
	}

	isAdmin := 0

	if data.IsAdmin {
		isAdmin = 1
	}

	return fmt.Sprintf("%d%s%s%s%s%s%d%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		data.Id,
		SeparateChar,
		data.Login,
		SeparateChar,
		data.Email,
		SeparateChar,
		isAdmin,
		SeparateChar,
		data.Sex.String,
		SeparateChar,
		data.Username,
		SeparateChar,
		avatar,
		SeparateChar,
		description,
		SeparateChar,
		signText,
		SeparateChar,
		createdAt,
		SeparateChar,
		updatedAt,
	)
}

func Deserialize(data string) (someInformation Account, err error) {
	toArray := strings.Split(data, SeparateChar)

	conv, err := strconv.Atoi(toArray[0])

	if err != nil {
		return someInformation, fmt.Errorf("%s: %w", "Deserialize", err)
	}

	var avatar sql.NullString
	var description sql.NullString
	var signText sql.NullString

	if toArray[6] != "" {
		avatar = sql.NullString{String: toArray[6], Valid: true}
	}

	if toArray[7] != "" {
		description = sql.NullString{String: toArray[7], Valid: true}
	}

	if toArray[8] != "" {
		signText = sql.NullString{String: toArray[8], Valid: true}
	}

	createdAtFormat, errT := time.Parse("2006-01-02 15:04:05", toArray[9])

	if errT != nil {
		log.Print("Deserailize account:", errT)
		createdAtFormat = time.Time{}
	}

	updatedAtFormat := sql.NullTime{}

	if toArray[10] != "" {
		updatedAtFormat.Time, _ = time.Parse(toArray[10], time.RFC3339)
		updatedAtFormat.Valid = true
	}

	isAdmin := false

	if toArray[3] == "1" {
		isAdmin = true
	}

	sex := sql.NullString{}

	if toArray[4] != "" {
		sex.String = toArray[4]
		sex.Valid = true
	}

	someInformation = Account{
		Id:          conv,
		Login:       toArray[1],
		Email:       toArray[2],
		IsAdmin:     isAdmin,
		Sex:         sex,
		Username:    toArray[5],
		Avatar:      avatar,
		Description: description,
		SignText:    signText,
		CreatedAt:   createdAtFormat,
		UpdatedAt:   updatedAtFormat,
	}

	return someInformation, nil
}
