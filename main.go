package main

import (
	"web-forum/api"
	"web-forum/frontend/web"
	"web-forum/internal"
)

func main() {
	internal.RegisterEnvironment()

	db := api.ConnectDatabase()
	rdb := api.ConnectToRedis()

	api.RegisterURLs(db, rdb)
	internal.SetDB(db)

	web.ExecuteNewServer(db, rdb)
}
