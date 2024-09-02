package main

import (
	"web-forum/system"
	"web-forum/system/redisDb"
	"web-forum/system/sqlDb"
	"web-forum/www"
)

func main() {
	system.RegisterEnvironment()

	db := sqlDb.ConnectDatabase()
	rdb := redisDb.ConnectToRedis()

	www.RegisterURLs(db, rdb)
}
