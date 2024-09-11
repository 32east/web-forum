package timer

import (
	"time"
	"web-forum/www/services/account"
)

func clearAccountsCache() {
	for id, object := range account.FastCache {
		// Почему не сразу очистить?
		// Потому что распределение нагрузки.

		if object.Time.Before(time.Now()) {
			delete(account.FastCache, id)
		}
	}
}

func Start() {
	timer := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-timer.C:
			clearAccountsCache()

			// ...
		}
	}
}
