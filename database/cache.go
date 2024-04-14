package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var CacheClient *redis.Client

func InitCache(config map[string]string) error {

	redisHost, ok := config["REDIS_HOST"]
	if !ok {
		return errors.New("REDIS_HOST env not found")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	status, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("Redis connection was refused", err)
	}
	fmt.Println(status)

	CacheClient = rdb

	return nil
}
