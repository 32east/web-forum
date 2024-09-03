package profile

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"
	"strings"
	"web-forum/api/auth"
	"web-forum/internal"
	"web-forum/system/sqlDb"
	"web-forum/www/services/account"
)

func HandleSettings(writer *http.ResponseWriter, reader *http.Request) {
	newJSONEncoder, answer := auth.PrepareHandle(writer)
	defer newJSONEncoder.Encode(answer)

	if reader.Method != "POST" {
		answer["success"], answer["reason"] = false, "method not allowed"

		return
	}

	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return
	}

	accountData, errGetAccount := account.ReadAccountFromCookie(cookie)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return
	}

	username := reader.FormValue("username")
	description := reader.FormValue("description")
	signText := reader.FormValue("signText")
	multipartFile, multiPartHeader, errFile := reader.FormFile("avatar")
	isAvatarRemove := reader.FormValue("avatarRemove") == "true"

	defer func() {
		if errFile != nil {
			return
		}

		multipartFile.Close()
	}()

	var valuesToChange = make(map[string]interface{})

	if username != "" {
		if len(username) < internal.UsernameMinLength {
			answer["success"], answer["reason"] = false, "username too short"
			return
		}

		valuesToChange["username"] = username
	}

	if description != "" {
		valuesToChange["description"] = description
	}

	if signText != "" {
		valuesToChange["sign_text"] = signText
	}

	if errFile == nil {
		contentTypeOfThisFile := multiPartHeader.Header["Content-Type"][0]

		if strings.Contains(contentTypeOfThisFile, "image/") == false {
			answer["success"], answer["reason"] = false, "file type not allowed"
			return
		}

		// Начинаем читать с 0 позиции.
		if _, seekErr := multipartFile.Seek(0, 0); seekErr != nil {
			answer["success"], answer["reason"] = false, "error seeking multipart file"
			return
		}

		config, format, decodeErr := image.Decode(multipartFile)

		if decodeErr != nil {
			answer["success"], answer["reason"] = false, "error decoding image"
			return
		}

		x, y := config.Bounds().Dx(), config.Bounds().Dy()
		x64, y64 := float64(x), float64(y)
		maxValue := math.Max(x64, y64)
		multiplier := internal.AvatarsSize / maxValue

		newWriter := new(bytes.Buffer)
		newImage := resize.Resize(uint(x64*multiplier), uint(y64*multiplier), config, resize.MitchellNetravali)

		switch format {
		case "jpeg":
			err = jpeg.Encode(newWriter, newImage, nil)
		case "png":
			err = png.Encode(newWriter, newImage)
		}

		if err != nil {
			answer["success"], answer["reason"] = false, "error encoding buffer to image"

			return
		}

		buf := make([]byte, newWriter.Len())
		_, readFileErr := newWriter.Read(buf)

		if readFileErr != nil {
			answer["success"], answer["reason"] = false, "error reading multipart file"

			return
		}

		newSha256Buffer := sha256.New()
		newSha256Buffer.Write(buf)
		encodeThisString := hex.EncodeToString(newSha256Buffer.Sum(nil))
		sixStartStr := encodeThisString[:6]

		fileName := fmt.Sprintf("%d-%s.%s", accountData.Id, sixStartStr, contentTypeOfThisFile[len("image/"):])

		if accountData.Avatar.Valid {
			os.Remove(internal.AvatarsFilePath + accountData.Avatar.String)
		}

		file, fileErr := os.Create(internal.AvatarsFilePath + fileName)
		defer file.Close()

		if fileErr != nil {
			answer["success"], answer["reason"] = false, "image is not uploaded, because file is not created"

			return
		}

		_, uploadFileError := file.Write(buf)

		if uploadFileError != nil {
			answer["success"], answer["reason"] = false, uploadFileError.Error()
			return
		}

		valuesToChange["avatar"] = fileName
	} else if isAvatarRemove {
		valuesToChange["avatar"] = nil
	}

	tx, err := sqlDb.MySqlDB.Begin()

	defer func(tx *sql.Tx) {
		if answer["success"] == false {
			tx.Rollback()
		}
	}(tx)

	if err != nil {
		answer["success"], answer["reason"] = false, err.Error()

		return
	}

	for key, value := range valuesToChange {
		formatQuery := fmt.Sprintf("UPDATE `users` SET `%s` = ? WHERE id = ?;", key)
		_, queryErr := tx.Exec(formatQuery, value, accountData.Id)

		if queryErr != nil {
			answer["success"], answer["reason"] = false, queryErr.Error()
			break
		}
	}

	transactionCommit := tx.Commit()

	if transactionCommit != nil {
		answer["success"], answer["reason"] = false, "transaction commit error"
		return
	}

	answer["success"] = true

	delete(account.CachedAccounts, accountData.Login)
	delete(account.CachedAccountsById, accountData.Id)
}
