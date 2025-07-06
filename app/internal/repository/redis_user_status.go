package repository

import (
	"context"
	"fmt"
	"time"

	"social-app/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisUserStatusRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisUserStatusRepository(client *redis.Client) domain.UserStatusRepository {
	return &RedisUserStatusRepository{
		client: client,
		ctx:    context.Background(),
	}
}

// SetUserStatus sets user status in Redis with TTL
func (r *RedisUserStatusRepository) SetUserStatus(userID, status string, ttl time.Duration) error {
	return r.setStatusWithBackup(userID, status, ttl)
}

// setStatusWithBackup sets status and maintains backup for auto-transition
func (r *RedisUserStatusRepository) setStatusWithBackup(userID, status string, ttl time.Duration) error {
	key := domain.GetUserStatusKey(userID)
	lastStatusKey := fmt.Sprintf("user:last_status:%s", userID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Set current status with TTL
	pipe.Set(r.ctx, key, status, ttl)

	// Set backup status (longer TTL for transition logic)
	backupTTL := ttl + (24 * time.Hour) // Keep backup longer than main key
	pipe.Set(r.ctx, lastStatusKey, status, backupTTL)

	_, err := pipe.Exec(r.ctx)
	return err
}

// GetUserStatus gets user status from Redis with auto-transition logic
func (r *RedisUserStatusRepository) GetUserStatus(userID string) (string, error) {
	key := domain.GetUserStatusKey(userID)
	result := r.client.Get(r.ctx, key)

	if result.Err() == redis.Nil {
		// Key doesn't exist - check for auto-transition
		return r.handleExpiredStatus(userID)
	}

	if result.Err() != nil {
		return "", result.Err()
	}

	status := result.Val()
	if status == "" {
		return domain.StatusUnknown, nil
	}

	return status, nil
}

// handleExpiredStatus implements auto-transition logic when keys expire
func (r *RedisUserStatusRepository) handleExpiredStatus(userID string) (string, error) {
	// Check if there's a backup status record to determine transition
	lastStatusKey := fmt.Sprintf("user:last_status:%s", userID)
	lastStatus, err := r.client.Get(r.ctx, lastStatusKey).Result()

	if err == redis.Nil || lastStatus == "" {
		// No previous status info - user is unknown
		return domain.StatusUnknown, nil
	}

	// Auto-transition based on last known status
	switch lastStatus {
	case domain.StatusOnline:
		// Online expired → Auto-transition to Away
		r.setStatusWithBackup(userID, domain.StatusAway, domain.AwayTTL)
		return domain.StatusAway, nil

	case domain.StatusAway:
		// Away expired → Auto-transition to Offline
		r.setStatusWithBackup(userID, domain.StatusOffline, domain.OfflineTTL)
		return domain.StatusOffline, nil

	case domain.StatusInvisible, domain.StatusDND:
		// Invisible/DND expired → Auto-transition to Away (they were active)
		r.setStatusWithBackup(userID, domain.StatusAway, domain.AwayTTL)
		return domain.StatusAway, nil

	default:
		// Offline or other status expired → Unknown
		return domain.StatusUnknown, nil
	}
}

// GetMultipleUserStatus gets multiple users status using MGET
func (r *RedisUserStatusRepository) GetMultipleUserStatus(userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return make(map[string]string), nil
	}

	// Build Redis keys
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = domain.GetUserStatusKey(userID)
	}

	// Execute MGET
	result := r.client.MGet(r.ctx, keys...)
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Process results
	statuses := make(map[string]string)
	values := result.Val()

	for i, userID := range userIDs {
		if i < len(values) && values[i] != nil {
			if status, ok := values[i].(string); ok && status != "" {
				statuses[userID] = status
			} else {
				statuses[userID] = domain.StatusUnknown
			}
		} else {
			statuses[userID] = domain.StatusUnknown
		}
	}

	return statuses, nil
}

// RefreshUserStatusTTL refreshes TTL for user status (heartbeat)
func (r *RedisUserStatusRepository) RefreshUserStatusTTL(userID string, ttl time.Duration) error {
	key := domain.GetUserStatusKey(userID)

	// Check if key exists and current status
	status, err := r.GetUserStatus(userID)
	if err != nil {
		return err
	}

	// Transition away→online when heartbeat received
	if status == domain.StatusAway {
		return r.client.Set(r.ctx, key, domain.StatusOnline, ttl).Err()
	}

	// Refresh TTL if user is online
	if status == domain.StatusOnline {
		return r.client.Expire(r.ctx, key, ttl).Err()
	}

	return nil
}
