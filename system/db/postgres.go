package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

var Postgres *pgxpool.Pool

func ConnectDatabase() *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	Postgres = conn
	return conn
}
