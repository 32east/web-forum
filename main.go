package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/system/timer"
	"web-forum/www"
)

var ctx = context.Background()

func main() {
	shutdownChan := make(chan os.Signal)
	signal.Notify(shutdownChan, os.Interrupt)

	err := os.Mkdir(internal.AvatarsFilePath, 0777)

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	system.RegisterEnvironment()

	conn := db.ConnectDatabase()
	defer conn.Close()

	redis := rdb.ConnectToRedis()
	defer redis.Close()

	// В кэше может быть устаревшая информация, например информация об аккаунтах.
	// При сбросе ДБ в Редисе сохранялся аккаунт, вследствие чего аккаунт вроде и был, а вроде и нет.
	redis.Do(ctx, "flushdb")

	go timer.Start()
	go www.RegisterURLs()

	<-shutdownChan
	log.Println("Сайт был закрыт.")
}
