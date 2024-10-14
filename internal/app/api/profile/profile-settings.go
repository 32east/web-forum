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
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/functions"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/account"
)

var ctx = context.Background()
var specialChars = "№^?=!@#$%^&*()_+ <>?:{}|;'/.,`~"

func containIllegalChars(str string) bool {
	var containIllegalCharacters = strings.IndexFunc(str, func(r rune) bool {
		if r >= 'А' && r <= 'я' {
			return false
		} else if r >= '0' && r <= '9' {
			return false
		}

		if strings.ContainsRune(specialChars, r) {
			return false
		}

		return r < 'A' || r > 'z'
	})

	return containIllegalCharacters >= 0
}

func ProcessSettings(accountData *account.Account, reader *http.Request, answer *map[string]interface{}) (err error) {
	var username = reader.FormValue("username")
	var sex = reader.FormValue("sex")
	var description = reader.FormValue("description")
	var signText = reader.FormValue("signText")
	var multipartFile, multiPartHeader, errFile = reader.FormFile("avatar")
	var isAvatarRemove = reader.FormValue("avatarRemove") == "true"

	if errFile == nil {
		defer multipartFile.Close()
	}

	var valuesToChange = make(map[string]interface{})

	if username != "" {
		var usernameLen = functions.Utf8Length(username)

		switch {
		case usernameLen < models.UsernameMinLength:
			(*answer)["success"], (*answer)["reason"] = false, "username too short"
			return nil
		case usernameLen > models.UsernameMaxLength:
			(*answer)["success"], (*answer)["reason"] = false, "username too long"
			return nil
		}

		username = functions.FormatString(username)

		var illegal = containIllegalChars(signText)

		if illegal {
			(*answer)["success"], (*answer)["reason"] = false, "illegal username"
			return nil
		}

		valuesToChange["username"] = username
	}

	if description != "" {
		description = functions.FormatString(description)
		description = strings.Replace(description, "\n\n", "\n", -1)

		if strings.Count(description, "\n") > 3 || functions.Utf8Length(description) > 512 {
			(*answer)["success"], (*answer)["reason"] = false, "description too long"

			return nil
		}

		var illegal = containIllegalChars(description)

		if illegal {
			(*answer)["success"], (*answer)["reason"] = false, "illegal description"
			return nil
		}

		valuesToChange["description"] = description
	} else {
		valuesToChange["description"] = nil
	}

	if sex != "" && (sex == "m" || sex == "f") {
		valuesToChange["sex"] = sex
	} else {
		valuesToChange["sex"] = nil
	}

	if signText != "" {
		signText = functions.FormatString(signText)
		var illegal = containIllegalChars(signText)

		if illegal {
			(*answer)["success"], (*answer)["reason"] = false, "illegal sign text"
			return nil
		}

		valuesToChange["sign_text"] = signText
	} else {
		valuesToChange["sign_text"] = nil
	}

	if errFile == nil {
		var contentTypeOfThisFile = multiPartHeader.Header["Content-Type"][0]

		if !strings.Contains(contentTypeOfThisFile, "image/") {
			(*answer)["success"], (*answer)["reason"] = false, "file type not allowed"
			return nil
		}

		// Начинаем читать с 0 позиции.
		if _, seekErr := multipartFile.Seek(0, 0); seekErr != nil {
			(*answer)["success"], (*answer)["reason"] = false, "error seeking multipart file"
			return seekErr
		}

		var config, format, decodeErr = image.Decode(multipartFile)

		if decodeErr != nil {
			(*answer)["success"], (*answer)["reason"] = false, "error decoding image"
			return decodeErr
		}

		var x, y = config.Bounds().Dx(), config.Bounds().Dy()
		var x64, y64 = float64(x), float64(y)
		var maxValue = math.Max(x64, y64)
		var multiplier = models.AvatarsSize / maxValue
		var newWriter = new(bytes.Buffer)
		var newImage = resize.Resize(uint(x64*multiplier), uint(y64*multiplier), config, resize.MitchellNetravali)

		switch format {
		case "jpeg":
			err = jpeg.Encode(newWriter, newImage, nil)
		case "png":
			err = png.Encode(newWriter, newImage)
		}

		if err != nil {
			(*answer)["success"], (*answer)["reason"] = false, "error encoding buffer to image"
			return err
		}

		var buf = make([]byte, newWriter.Len())
		var _, readFileErr = newWriter.Read(buf)

		if readFileErr != nil {
			(*answer)["success"], (*answer)["reason"] = false, "error reading multipart file"
			return readFileErr
		}

		var newSha256Buffer = sha256.New()
		newSha256Buffer.Write(buf)
		var encodeThisString = hex.EncodeToString(newSha256Buffer.Sum(nil))
		var sixStartStr = encodeThisString[:6]
		var fileName = fmt.Sprintf("%d-%s.%s", (*accountData).Id, sixStartStr, contentTypeOfThisFile[len("image/"):])

		if accountData.Avatar.Valid {
			os.Remove(models.AvatarsFilePath + (*accountData).Avatar.String)
		}

		var file, fileErr = os.Create(models.AvatarsFilePath + fileName)

		if fileErr != nil {
			(*answer)["success"], (*answer)["reason"] = false, "image is not uploaded, because file is not created"
			return fileErr
		}
		defer file.Close()

		var _, uploadFileError = file.Write(buf)

		if uploadFileError != nil {
			(*answer)["success"], (*answer)["reason"] = false, "const-funcs server error"
			return uploadFileError
		}

		valuesToChange["avatar"] = fileName
	} else if isAvatarRemove {
		valuesToChange["avatar"] = nil
	}

	var tx, txErr = db.Postgres.Begin(ctx)

	if txErr != nil {
		(*answer)["success"], (*answer)["reason"] = false, "const-funcs server error"
		return txErr
	}

	defer func() {
		switch (*answer)["success"] {
		case true:
			tx.Commit(ctx)
		case false:
			tx.Rollback(ctx)
		}
	}()

	// Может потом переписать??
	for key, value := range valuesToChange {
		var formatQuery = fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2;", key)
		var _, queryErr = tx.Exec(ctx, formatQuery, value, (*accountData).Id)

		if queryErr != nil {
			(*answer)["success"], (*answer)["reason"] = false, "const-funcs server error"
			return queryErr
		}
	}

	(*answer)["success"] = true

	rdb.RedisDB.Del(ctx, fmt.Sprintf("aID:%d", (*accountData).Id))
	delete(account.FastCache, (*accountData).Id)
	return nil
}

func HandleSettings(_ http.ResponseWriter, reader *http.Request, answer map[string]interface{}) error {
	var cookie, err = reader.Cookie("access_token")

	if err != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	var accountData, errGetAccount = account.ReadFromCookie(cookie)

	if errGetAccount != nil {
		answer["success"], answer["reason"] = false, "not authorized"
		return nil
	}

	return ProcessSettings(accountData, reader, &answer)
}
