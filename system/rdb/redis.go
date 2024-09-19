package rdb

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
	"time"
)

var RedisDB *redis.Client
var ctx = context.Background()
var failsCount = 1 // Зарезервировано.

func TryToConnect() *redis.Client {
	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       redisDb,
	})

	rdbErr := rdb.Ping(ctx).Err()

	if rdbErr == nil {
		return rdb
	}

	return nil
}

func ConnectToRedis() *redis.Client {
	RedisDB = TryToConnect()

	if RedisDB != nil {
		return RedisDB
	}

	newTicker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-newTicker.C:
			RedisDB = TryToConnect()

			if RedisDB != nil {
				return RedisDB
			} else {
				failsCount += 1

				if failsCount >= 5 {
					panic("failed to connect to redis after 5 attempts")
				}
			}
		}
	}
}
