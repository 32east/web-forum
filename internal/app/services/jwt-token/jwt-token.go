package jwt_token

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"strconv"
	"time"
	"web-forum/internal/app/models"
)

func GenerateNew[T int | int64](accountId T, additionalParam string) (string, error) {
	var expirationTime = time.Now().Add(time.Hour * 24)

	if additionalParam == "refresh" {
		expirationTime = time.Now().Add(time.Hour * 72)
	}

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         accountId,
		"additional": additionalParam,
		"expiresAt":  fmt.Sprintf("%d", expirationTime.Unix()),
	})

	return token.SignedString(models.HmacSecret)
}

func GetInfo(token string) (jwt.MapClaims, error) {
	var tokenParse, parseErr = jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return models.HmacSecret, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	var tokenClaim = tokenParse.Claims.(jwt.MapClaims)
	var id, idOk = tokenClaim["id"].(float64)

	if !idOk {
		return nil, fmt.Errorf("invalid id")
	}

	tokenClaim["id"] = int64(id)
	var exp, ok = tokenClaim["expiresAt"].(string)

	if !ok {
		return nil, fmt.Errorf("invalid exp")
	}

	var expInt64, parseIntErr = strconv.ParseInt(exp, 10, 64)

	if parseIntErr != nil {
		return nil, parseIntErr
	}

	if expInt64 < time.Now().Unix() {
		return nil, fmt.Errorf("token is expired")
	}

	return tokenClaim, nil
}
