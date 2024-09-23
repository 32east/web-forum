package profile

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"web-forum/internal"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/www/services/account"
)

var ctx = context.Background()

func HandleSettings(writer http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	cookie, err := reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	accountData, errGetAccount := account.ReadFromCookie(cookie)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	username := reader.FormValue("username")
	description := reader.FormValue("description")
	signText := reader.FormValue("signText")
	multipartFile, multiPartHeader, errFile := reader.FormFile("avatar")
	isAvatarRemove := reader.FormValue("avatarRemove") == "true"

	if errFile != nil {
		defer multipartFile.Close()
	}

	var valuesToChange = make(map[string]interface{})

	if username != "" {
		usernameLen := internal.Utf8Length(username)

		switch {
		case usernameLen < internal.UsernameMinLength:
			answer["success"], answer["reason"] = false, "username too short"
			return nil
		case usernameLen > internal.UsernameMaxLength:
			answer["success"], answer["reason"] = false, "username too long"
			return nil
		}

		username = internal.FormatString(username)

		containIllegalCharacters := strings.IndexFunc(username, func(r rune) bool {
			if r >= 'А' && r <= 'я' {
				return false
			} else if r >= '0' && r <= '9' {
				return false
			}

			return r < 'A' || r > 'z'
		})

		if containIllegalCharacters >= 0 {
			answer["success"], answer["reason"] = false, "username contains illegal characters"
			return nil
		}

		valuesToChange["username"] = username
	}

	if description != "" {
		description = internal.FormatString(description)
		description = strings.Replace(description, "\n\n", "\n", -1)

		if strings.Count(description, "\n") > 3 || internal.Utf8Length(description) > 512 {
			answer["success"], answer["reason"] = false, "description too long"

			return nil
		}

		valuesToChange["description"] = description
	} else {
		valuesToChange["description"] = nil
	}

	if signText != "" {
		signText = internal.FormatString(signText)
		valuesToChange["sign_text"] = signText
	} else {
		valuesToChange["sign_text"] = nil
	}

	if errFile == nil {
		contentTypeOfThisFile := multiPartHeader.Header["Content-Type"][0]

		if !strings.Contains(contentTypeOfThisFile, "image/") {
			answer["success"], answer["reason"] = false, "file type not allowed"
			return nil
		}

		// Начинаем читать с 0 позиции.
		if _, seekErr := multipartFile.Seek(0, 0); seekErr != nil {
			answer["success"], answer["reason"] = false, "error seeking multipart file"
			return seekErr
		}

		config, format, decodeErr := image.Decode(multipartFile)

		if decodeErr != nil {
			answer["success"], answer["reason"] = false, "error decoding image"
			return decodeErr
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
			return err
		}

		buf := make([]byte, newWriter.Len())
		_, readFileErr := newWriter.Read(buf)

		if readFileErr != nil {
			answer["success"], answer["reason"] = false, "error reading multipart file"
			return readFileErr
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
		if file != nil {
			defer file.Close()
		}

		if fileErr != nil {
			answer["success"], answer["reason"] = false, "image is not uploaded, because file is not created"
			return fileErr
		}

		_, uploadFileError := file.Write(buf)

		if uploadFileError != nil {
			answer["success"], answer["reason"] = false, "internal server error"
			return uploadFileError
		}

		valuesToChange["avatar"] = fileName
	} else if isAvatarRemove {
		valuesToChange["avatar"] = nil
	}

	tx, err := db.Postgres.Begin(ctx)

	defer func() {
		switch answer["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	if err != nil {
		answer["success"], answer["reason"] = false, "internal server error"
		return err
	}

	// Может потом переписать??
	for key, value := range valuesToChange {
		formatQuery := fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2;", key)
		_, queryErr := tx.Exec(ctx, formatQuery, value, accountData.Id)

		if queryErr != nil {
			answer["success"], answer["reason"] = false, "internal server error"
			return queryErr
		}
	}

	answer["success"] = true

	rdb.RedisDB.Del(ctx, fmt.Sprintf("aID:%d", accountData.Id))
	delete(account.FastCache, accountData.Id)
	return nil
}
