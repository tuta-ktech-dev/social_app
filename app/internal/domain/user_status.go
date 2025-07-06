package domain

import "time"

// User status constants
const (
	StatusOnline  = "online"
	StatusOffline = "offline"
)

// Redis key patterns
const (
	UserStatusKeyPrefix = "user:status:"
	DefaultTTL          = 30 * time.Second
)

// UserStatus represents user online/offline status
type UserStatus struct {
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// UserStatusRepository interface for Redis operations
type UserStatusRepository interface {
	SetUserStatus(userID, status string, ttl time.Duration) error
	GetUserStatus(userID string) (string, error)
	GetMultipleUserStatus(userIDs []string) (map[string]string, error)
	RefreshUserStatusTTL(userID string, ttl time.Duration) error
}

// GetRedisKey returns Redis key for user status
func GetUserStatusKey(userID string) string {
	return UserStatusKeyPrefix + userID
} 