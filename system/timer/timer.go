package timer

import (
	"context"
	"time"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/www/services/account"
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
	_, err := db.Postgres.Exec(ctx, "delete from tokens where expires_at < now();")

	if err != nil {
		system.ErrLog("timer.clearRefreshTokens", err)
	}
}

func Start() {
	timer := time.NewTicker(time.Second * 10)
	timerHour := time.NewTicker(time.Hour)

	for {
		select {
		case <-timer.C:
			clearAccountsCache()

		case <-timerHour.C:
			clearRefreshTokens()
		}
	}
}
