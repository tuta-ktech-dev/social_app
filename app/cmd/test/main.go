package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"social-app/config"
	"social-app/internal/domain"
	"social-app/internal/repository"
)

func main() {
	fmt.Println("=== Redis User Status Test Runner ===")

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

	// Initialize user status repository
	userStatusRepo := repository.NewRedisUserStatusRepository(redisClient)

	// Run comprehensive tests
	runUserStatusTests(userStatusRepo)

	fmt.Println("\n=== All tests completed ===")
}

func runUserStatusTests(repo domain.UserStatusRepository) {
	// Test 1: Set user online
	fmt.Println("\n1. Setting user 123 online...")
	err := repo.SetUserStatus("123", domain.StatusOnline, domain.DefaultTTL)
	if err != nil {
		log.Printf("❌ Error setting user online: %v", err)
	} else {
		fmt.Println("✅ User 123 set to online")
	}

	// Test 2: Get user status
	fmt.Println("\n2. Getting user 123 status...")
	status, err := repo.GetUserStatus("123")
	if err != nil {
		log.Printf("❌ Error getting user status: %v", err)
	} else {
		fmt.Printf("✅ User 123 status: %s\n", status)
	}

	// Test 3: Set multiple users with different statuses
	fmt.Println("\n3. Setting multiple users status...")
	repo.SetUserStatus("456", domain.StatusOnline, domain.DefaultTTL)
	repo.SetUserStatus("789", domain.StatusOffline, 24*time.Hour)
	fmt.Println("✅ Set user 456 online, user 789 offline")

	// Test 4: Get multiple users status
	fmt.Println("\n4. Getting multiple users status...")
	userIDs := []string{"123", "456", "789", "999"}
	statuses, err := repo.GetMultipleUserStatus(userIDs)
	if err != nil {
		log.Printf("❌ Error getting multiple user status: %v", err)
	} else {
		fmt.Println("✅ Multiple users status:")
		for userID, status := range statuses {
			fmt.Printf("   - User %s: %s\n", userID, status)
		}
	}

	// Test 5: Refresh TTL (heartbeat)
	fmt.Println("\n5. Refreshing user 123 TTL (heartbeat)...")
	err = repo.RefreshUserStatusTTL("123", domain.DefaultTTL)
	if err != nil {
		log.Printf("❌ Error refreshing TTL: %v", err)
	} else {
		fmt.Println("✅ User 123 TTL refreshed")
	}

	// Test 6: Wait and check TTL behavior
	fmt.Println("\n6. Waiting 5 seconds to test TTL...")
	time.Sleep(5 * time.Second)

	status, err = repo.GetUserStatus("123")
	if err != nil {
		log.Printf("❌ Error getting user status after wait: %v", err)
	} else {
		fmt.Printf("✅ User 123 status after 5s: %s\n", status)
	}

	// Test 7: Test non-existent user
	fmt.Println("\n7. Testing non-existent user...")
	status, err = repo.GetUserStatus("nonexistent")
	if err != nil {
		log.Printf("❌ Error getting non-existent user: %v", err)
	} else {
		fmt.Printf("✅ Non-existent user status: %s\n", status)
	}
}
