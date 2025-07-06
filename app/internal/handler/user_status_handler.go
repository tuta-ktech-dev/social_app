package handler

import (
	"fmt"
	"net/http"
	"social-app/internal/domain"
	"social-app/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

// Request DTOs
type SetStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type HeartbeatRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// Response DTOs
type UserStatusResponse struct {
	Success bool               `json:"success"`
	Data    *domain.UserStatus `json:"data,omitempty"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
}

type MultipleUserStatusResponse struct {
	Success bool                          `json:"success"`
	Data    map[string]*domain.UserStatus `json:"data,omitempty"`
	Count   int                           `json:"count,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

type HeartbeatResponse struct {
	Success              bool   `json:"success"`
	Message              string `json:"message,omitempty"`
	NextHeartbeatSeconds int    `json:"next_heartbeat_seconds,omitempty"`
	Error                string `json:"error,omitempty"`
}

type UserStatusHandler struct {
	service *services.UserStatusService
}

func NewUserStatusHandler(service *services.UserStatusService) *UserStatusHandler {
	return &UserStatusHandler{service: service}
}

// POST /users/:id/status
// Set user status (online/away/offline)
func (h *UserStatusHandler) SetUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate status value
	if !isValidStatus(req.Status) {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   "Invalid status. Must be one of: online, away, offline, invisible, dnd",
		})
		return
	}

	if err := h.service.SetUserStatus(userID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data: &domain.UserStatus{
			UserID:    userID,
			Status:    req.Status,
			Timestamp: time.Now(),
		},
		Message: "User status updated successfully",
	})
}

// GET /users/:id/status
// Get user status
func (h *UserStatusHandler) GetUserStatus(c *gin.Context) {
	userID := c.Param("id")

	status, err := h.service.GetUserStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data:    status,
	})
}

// GET /users/status?user_ids=123,456,789
// Get multiple users status
func (h *UserStatusHandler) GetMultipleUserStatus(c *gin.Context) {
	userIDs := c.QueryArray("user_ids")

	if len(userIDs) == 0 {
		c.JSON(http.StatusBadRequest, MultipleUserStatusResponse{
			Success: false,
			Error:   "user_ids parameter is required",
		})
		return
	}

	statuses, err := h.service.GetMultipleUserStatus(userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MultipleUserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MultipleUserStatusResponse{
		Success: true,
		Data:    statuses,
		Count:   len(statuses),
	})
}

// POST /users/:id/heartbeat
// Send heartbeat to maintain online status
func (h *UserStatusHandler) SendHeartbeat(c *gin.Context) {
	userID := c.Param("id")

	if err := h.service.SendHeartbeat(userID); err != nil {
		c.JSON(http.StatusBadRequest, HeartbeatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, HeartbeatResponse{
		Success:              true,
		Message:              "Heartbeat received successfully",
		NextHeartbeatSeconds: 30, // Online TTL duration
	})
}

// PUT /users/:id/status/away
// Set user to away status
func (h *UserStatusHandler) SetUserAway(c *gin.Context) {
	userID := c.Param("id")

	if err := h.service.SetUserAway(userID); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data: &domain.UserStatus{
			UserID:    userID,
			Status:    domain.StatusAway,
			Timestamp: time.Now(),
		},
		Message: "User set to away status",
	})
}

// PUT /users/:id/status/offline
// Set user to offline status
func (h *UserStatusHandler) SetUserOffline(c *gin.Context) {
	userID := c.Param("id")

	if err := h.service.SetUserOffline(userID); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data: &domain.UserStatus{
			UserID:    userID,
			Status:    domain.StatusOffline,
			Timestamp: time.Now(),
		},
		Message: "User set to offline status",
	})
}

// PUT /users/:id/status/invisible
// Set user to invisible status
func (h *UserStatusHandler) SetUserInvisible(c *gin.Context) {
	userID := c.Param("id")

	if err := h.service.SetUserInvisible(userID); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data: &domain.UserStatus{
			UserID:    userID,
			Status:    domain.StatusInvisible,
			Timestamp: time.Now(),
		},
		Message: "User set to invisible status",
	})
}

// PUT /users/:id/status/dnd
// Set user to do not disturb status
func (h *UserStatusHandler) SetUserDND(c *gin.Context) {
	userID := c.Param("id")
	fmt.Printf("🔍 [DEBUG] SetUserDND - Received userID: '%s'\n", userID)

	if err := h.service.SetUserDND(userID); err != nil {
		c.JSON(http.StatusBadRequest, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data: &domain.UserStatus{
			UserID:    userID,
			Status:    domain.StatusDND,
			Timestamp: time.Now(),
		},
		Message: "User set to do not disturb status",
	})
}

// GET /users/:id/status/public
// Get user status visible to other users (handles invisible mode)
func (h *UserStatusHandler) GetPublicUserStatus(c *gin.Context) {
	userID := c.Param("id")

	status, err := h.service.GetPublicUserStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserStatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserStatusResponse{
		Success: true,
		Data:    status,
	})
}

// Helper function to validate status
func isValidStatus(status string) bool {
	switch status {
	case domain.StatusOnline, domain.StatusAway, domain.StatusOffline, domain.StatusInvisible, domain.StatusDND:
		return true
	default:
		return false
	}
}
