package jwt_token

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"strconv"
	"time"
	"web-forum/internal"
)

func GetTokenInfo(token string) (jwt.MapClaims, error) {
	tokenParse, parseErr := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return internal.HmacSecret, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	tokenClaim := tokenParse.Claims.(jwt.MapClaims)
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
