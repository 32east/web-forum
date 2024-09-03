package main

import (
	"web-forum/system"
	"web-forum/system/redisDb"
	"web-forum/system/sqlDb"
	"web-forum/www"
)

// TODO: Позже сделать обработчик, что если access_token уже устаревший, то гляди на refresh_token.
// TODO: Если и refresh_token устаревший, то всё пизда.

func main() {
	system.RegisterEnvironment()

	sqlDb.ConnectDatabase()
	redisDb.ConnectToRedis()

	www.RegisterURLs()
}
