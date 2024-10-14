package timer

import (
	"context"
	"time"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/services/account"
	"web-forum/pkg/stuff"
)

var ctx = context.Background()

func clearAccountsCache() {
	for id, object := range account.FastCache {
		// Почему не сразу очистить?
		// Потому что распределение нагрузки.

		if object.Time.Before(time.Now()) {
			delete(account.FastCache, id)
		}
	}
}

func clearRefreshTokens() {
	var _, err = db.Postgres.Exec(ctx, "delete from tokens where expiresat < now();")

	if err != nil {
		stuff.ErrLog("timer.clearRefreshTokens", err)
	}
}

func Start() {
	var timer = time.NewTicker(time.Second * 10)
	var timerHour = time.NewTicker(time.Hour)

	for {
		select {
		case <-timer.C:
			clearAccountsCache()

		case <-timerHour.C:
			clearRefreshTokens()
		}
	}
}
