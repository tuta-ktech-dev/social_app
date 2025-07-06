package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379", // redis container name in docker-compose
	})

	// Test set + get
	err := rdb.Set(ctx, "user:online:123", "1", 30*time.Second).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "user:online:123").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("user online status:", val)
}
