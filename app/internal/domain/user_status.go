package domain

import "time"

// User status constants
const (
	StatusOnline    = "online"
	StatusAway      = "away"
	StatusOffline   = "offline"
	StatusInvisible = "invisible"
	StatusDND       = "dnd"
	StatusUnknown   = "unknown"
)

// Redis key patterns and TTL values
const (
	UserStatusKeyPrefix = "user:status:"
	OnlineTTL           = 30 * time.Second
	AwayTTL             = 10 * time.Minute
	OfflineTTL          = 24 * time.Hour
)

// UserStatus represents user online/offline status
type UserStatus struct {
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	ActualStatus string    `json:"actual_status,omitempty"` // For invisible mode
	Timestamp    time.Time `json:"timestamp"`
	LastActivity time.Time `json:"last_activity,omitempty"`
}

// NotificationPreference represents user's notification settings
type NotificationPreference struct {
	UserID              string     `json:"user_id"`
	AllowDirectMessages bool       `json:"allow_direct_messages"`
	AllowMentions       bool       `json:"allow_mentions"`
	AllowGroupMessages  bool       `json:"allow_group_messages"`
	DNDUntil            *time.Time `json:"dnd_until,omitempty"`
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
