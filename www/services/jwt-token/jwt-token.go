package jwt_token

import (
	"crypto/rand"
	"fmt"
	"github.com/golang-jwt/jwt"
	"strconv"
	"time"
	"web-forum/internal"
)

func GenerateNew[T int | int64](accountId T, additionalParam string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(time.Hour * 24)

	if additionalParam == "refresh" {
		expirationTime = time.Now().Add(time.Hour * 72)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         accountId,
		"additional": additionalParam,
		"expiresAt":  fmt.Sprintf("%d", expirationTime.Unix()),
	})

	tokenStr, err := token.SignedString(internal.HmacSecret)

	return tokenStr, err
}

func GetInfo(token string) (jwt.MapClaims, error) {
	tokenParse, parseErr := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return internal.HmacSecret, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	tokenClaim := tokenParse.Claims.(jwt.MapClaims)
	id, idOk := tokenClaim["id"].(float64)

	if !idOk {
		return nil, fmt.Errorf("invalid id")
	}

	tokenClaim["id"] = int64(id)
	exp, ok := tokenClaim["expiresAt"].(string)

	if !ok {
		return nil, fmt.Errorf("invalid exp")
	}

	expInt64, parseIntErr := strconv.ParseInt(exp, 10, 64)

	if parseIntErr != nil {
		return nil, parseIntErr
	}

	if expInt64 < time.Now().Unix() {
		return nil, fmt.Errorf("token is expired")
	}

	return tokenClaim, nil
}
