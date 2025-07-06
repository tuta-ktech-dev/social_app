package main

import (
	"context"
	"fmt"
	"log"

	"social-app/config"
)

func main() {
	fmt.Println("Social App Starting...")

	// Initialize Redis configuration
	redisConfig := config.NewRedisConfig()
	redisClient := redisConfig.NewClient()

	// Test Redis connection
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	fmt.Println("Redis connected:", pong)

	fmt.Println("\nSocial App running successfully!")
}
