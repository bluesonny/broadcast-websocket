package models

import (
	. "broadcast-websocket/config"
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var RedisClient *redis.Client
var ctx = context.Background()

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     ViperConfig.Redis.Address,
		Password: ViperConfig.Redis.Auth, // no password set
		DB:       0,                      // use default DB
	})
	pong, err := RedisClient.Ping(ctx).Result()
	//fmt.Println("Redis Client: " + pong)
	if err == nil {
		log.Println("redis 正常工作...")
	} else {
		log.Printf("Redis Client err %v\n", err)
	}
	log.Printf("redis ping result: %s\n", pong)
}
