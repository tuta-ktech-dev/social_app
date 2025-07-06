package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"social-app/config"
	"social-app/internal/repository"
	"social-app/internal/router"
	"social-app/internal/services"
)

func main() {
	fmt.Println("🔥 HOT RELOAD IS WORKING! 🔥")

	// Initialize Redis configuration
	redisConfig := config.NewRedisConfig()
	redisClient := redisConfig.NewClient()

	// Test Redis connection
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}
	fmt.Println("✅ Redis connected:", pong)

	// Initialize dependencies
	userStatusRepo := repository.NewRedisUserStatusRepository(redisClient)
	userStatusService := services.NewUserStatusService(userStatusRepo)

	// Setup router
	r := router.SetupRouter(userStatusService)

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("🌐 Starting HTTP server on port 8080...")
		fmt.Println("📚 API Documentation: http://localhost:8080")
		fmt.Println("🔍 Health Check: http://localhost:8080/health")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	fmt.Println("✅ Server exited gracefully")
}
