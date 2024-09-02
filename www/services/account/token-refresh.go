package account

import (
	"context"
	"net/http"
	"time"
	"web-forum/system/redisDb"
)

func TokensRefreshInRedis(reader *http.Request, writer *http.ResponseWriter) {
	rdb := redisDb.RedisDB

	if reader.Referer() != "" {
		return
	}

	ctx := context.Background()
	accessToken, accessCookieErr := reader.Cookie("access_token")

	if accessCookieErr != nil {
		return
	}

	refreshToken, errRefresh := reader.Cookie("refresh_token")

	if errRefresh != nil {
		return
	}

	resultAccessToken, errAccessToken := rdb.Get(ctx, "AToken:"+accessToken.Value).Result()
	resultRefreshToken, errRefreshToken := rdb.Get(ctx, "RToken:"+refreshToken.Value).Result()

	if errAccessToken == nil {
		rdb.Set(ctx, "AToken:"+accessToken.Value, resultAccessToken, time.Hour*12)
	}

	if errRefreshToken == nil {
		rdb.Set(ctx, "RToken:"+refreshToken.Value, resultRefreshToken, time.Hour*72)
	}

	http.SetCookie(*writer, &http.Cookie{
		Name:    "access_token",
		Value:   accessToken.Value,
		Expires: time.Now().Add(time.Hour * 12),
		Path:    "/",
	})

	http.SetCookie(*writer, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken.Value,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 72),
	})
}
