package models

import (
	. "broadcast-websocket/config"
	"fmt"
	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     ViperConfig.Redis.Address,
		Password: ViperConfig.Redis.Auth, // no password set
		DB:       0,                      // use default DB
	})
	_, err := RedisClient.Ping().Result()
	//fmt.Println("Redis Client: " + pong)
	if err == nil {
		fmt.Println("redis 正常工作...")
	} else {
		fmt.Printf("Redis Client err %v\n", err)
	}
}
