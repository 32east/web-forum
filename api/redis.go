package api

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
)

func ConnectToRedis() *redis.Client {
	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       redisDb,
	})

	rdbErr := rdb.Ping(context.Background()).Err()

	if rdbErr != nil {
		panic(rdbErr)
	}

	return rdb
}
