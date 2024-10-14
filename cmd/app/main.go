package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/models"
	"web-forum/internal/app/timer"
	"web-forum/internal/app/transport"
	"web-forum/pkg/stuff"
)

var ctx = context.Background()

func main() {
	shutdownChan := make(chan os.Signal)
	signal.Notify(shutdownChan, os.Interrupt)

	err := os.Mkdir(models.AvatarsFilePath, 0777)

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	stuff.RegisterEnvironment()

	conn := db.ConnectDatabase()
	defer conn.Close()

	redis := rdb.ConnectToRedis()
	defer redis.Close()

	// В кэше может быть устаревшая информация, например информация об аккаунтах.
	// При сбросе ДБ в Редисе сохранялся аккаунт, вследствие чего аккаунт вроде и был, а вроде и нет.
	redis.Do(ctx, "flushdb")

	go timer.Start()
	go transport.RegisterURLs()

	<-shutdownChan
	log.Println("Сайт был закрыт.")
}
