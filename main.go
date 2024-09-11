package main

import (
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/system/timer"
	"web-forum/www"
)

// TODO: Позже сделать обработчик, что если access_token уже устаревший, то гляди на refresh_token.
// TODO: Если и refresh_token устаревший, то всё пизда.

func main() {
	system.RegisterEnvironment()

	db.ConnectDatabase()
	rdb.ConnectToRedis()

	go timer.Start()

	www.RegisterURLs()
}
