package router

import (
	"social-app/internal/handler"
	"social-app/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter(userStatusService *services.UserStatusService) *gin.Engine {
	// Create Gin router
	r := gin.Default()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Initialize handlers
	userStatusHandler := handler.NewUserStatusHandler(userStatusService)

	// API versioning
	v1 := r.Group("/api/v1")
	{
		// User status routes
		users := v1.Group("/users")
		{
			// Individual user status
			users.POST("/:id/status", userStatusHandler.SetUserStatus)             // Set status (online/away/offline/invisible/dnd)
			users.GET("/:id/status", userStatusHandler.GetUserStatus)              // Get user status (own status)
			users.GET("/:id/status/public", userStatusHandler.GetPublicUserStatus) // Get public status (visible to others)
			users.POST("/:id/heartbeat", userStatusHandler.SendHeartbeat)          // Send heartbeat
			users.PUT("/:id/status/away", userStatusHandler.SetUserAway)           // Set user away
			users.PUT("/:id/status/offline", userStatusHandler.SetUserOffline)     // Set user offline
			users.PUT("/:id/status/invisible", userStatusHandler.SetUserInvisible) // Set user invisible
			users.PUT("/:id/status/dnd", userStatusHandler.SetUserDND)             // Set user do not disturb

			// Bulk operations
			users.GET("/status", userStatusHandler.GetMultipleUserStatus) // Get multiple users status
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Social App API is running",
		})
	})

	// API documentation endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Social App API",
			"version": "1.0.0",
			"endpoints": map[string]interface{}{
				"health": "GET /health",
				"user_status": map[string]string{
					"set_status":        "POST /api/v1/users/:id/status",
					"get_status":        "GET /api/v1/users/:id/status",
					"get_public_status": "GET /api/v1/users/:id/status/public",
					"send_heartbeat":    "POST /api/v1/users/:id/heartbeat",
					"set_away":          "PUT /api/v1/users/:id/status/away",
					"set_offline":       "PUT /api/v1/users/:id/status/offline",
					"set_invisible":     "PUT /api/v1/users/:id/status/invisible",
					"set_dnd":           "PUT /api/v1/users/:id/status/dnd",
					"get_multiple":      "GET /api/v1/users/status?user_ids=123,456",
				},
			},
		})
	})

	return r
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
