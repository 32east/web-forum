package main

import (
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/system/timer"
	"web-forum/www"
)

func main() {
	system.RegisterEnvironment()

	conn := db.ConnectDatabase()
	defer conn.Close()

	redis := rdb.ConnectToRedis()
	defer redis.Close()

	go timer.Start()

	www.RegisterURLs()
}
