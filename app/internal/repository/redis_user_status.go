package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"social-app/internal/domain"
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
	key := domain.GetUserStatusKey(userID)
	return r.client.Set(r.ctx, key, status, ttl).Err()
}

// GetUserStatus gets user status from Redis
func (r *RedisUserStatusRepository) GetUserStatus(userID string) (string, error) {
	key := domain.GetUserStatusKey(userID)
	result := r.client.Get(r.ctx, key)
	
	if result.Err() == redis.Nil {
		// Key doesn't exist, user is offline
		return domain.StatusOffline, nil
	}
	
	if result.Err() != nil {
		return "", result.Err()
	}
	
	status := result.Val()
	if status == "" {
		return domain.StatusOffline, nil
	}
	
	return status, nil
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
				statuses[userID] = domain.StatusOffline
			}
		} else {
			statuses[userID] = domain.StatusOffline
		}
	}
	
	return statuses, nil
}

// RefreshUserStatusTTL refreshes TTL for user status (heartbeat)
func (r *RedisUserStatusRepository) RefreshUserStatusTTL(userID string, ttl time.Duration) error {
	key := domain.GetUserStatusKey(userID)
	
	// Check if key exists and has "online" status
	status, err := r.GetUserStatus(userID)
	if err != nil {
		return err
	}
	
	// Only refresh TTL if user is online
	if status == domain.StatusOnline {
		return r.client.Expire(r.ctx, key, ttl).Err()
	}
	
	return nil
} 