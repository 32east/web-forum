package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

var Postgres *pgxpool.Pool
var ctx = context.Background()
var failsCount = 1 // Зарезервировано.

func TryToConnect() *pgxpool.Pool {
	conn, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Println(err)
		return nil
	}

	pingErr := conn.Ping(ctx)

	if pingErr != nil {
		log.Println(pingErr)
		return nil
	}

	log.Println("Successfully connected to database")

	return conn
}

func ConnectDatabase() *pgxpool.Pool {
	Postgres = TryToConnect()

	if Postgres != nil {
		return Postgres
	}

	newTicker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-newTicker.C:
			Postgres = TryToConnect()

			if Postgres != nil {
				return Postgres
			} else {
				failsCount += 1

				if failsCount >= 5 {
					panic("failed to connect to database after 5 attempts")
				}
			}
		}
	}
}
