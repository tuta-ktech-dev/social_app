package services

import (
	"errors"
	"social-app/internal/domain"
	"strings"
	"time"
)

type UserStatusService struct {
	repo domain.UserStatusRepository
}

// NewUserStatusService creates a new UserStatusService with the given repository
func NewUserStatusService(repo domain.UserStatusRepository) *UserStatusService {
	return &UserStatusService{
		repo: repo,
	}
}

// Business methods
func (s *UserStatusService) SetUserStatus(userID string, status string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}

	// Determine appropriate TTL based on status
	var ttl time.Duration
	switch status {
	case domain.StatusOnline:
		ttl = domain.OnlineTTL
	case domain.StatusAway:
		ttl = domain.AwayTTL
	case domain.StatusOffline:
		ttl = domain.OfflineTTL
	case domain.StatusInvisible:
		ttl = domain.OnlineTTL // Same as online but appears offline
	case domain.StatusDND:
		ttl = domain.OnlineTTL // Same as online but limited notifications
	default:
		return errors.New("invalid status: must be online, away, offline, invisible, or dnd")
	}

	return s.repo.SetUserStatus(userID, status, ttl)
}

func (s *UserStatusService) SetUserOffline(userID string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}
	return s.repo.SetUserStatus(userID, domain.StatusOffline, domain.OfflineTTL)
}

func (s *UserStatusService) SetUserAway(userID string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}
	return s.repo.SetUserStatus(userID, domain.StatusAway, domain.AwayTTL)
}

func (s *UserStatusService) SetUserInvisible(userID string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}
	return s.repo.SetUserStatus(userID, domain.StatusInvisible, domain.OnlineTTL)
}

func (s *UserStatusService) SetUserDND(userID string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}
	return s.repo.SetUserStatus(userID, domain.StatusDND, domain.OnlineTTL)
}

func (s *UserStatusService) GetUserStatus(userID string) (*domain.UserStatus, error) {
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	status, err := s.repo.GetUserStatus(userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserStatus{
		UserID:    userID,
		Status:    status,
		Timestamp: time.Now(),
	}, nil
}

// GetPublicUserStatus returns status visible to other users (handles invisible mode)
func (s *UserStatusService) GetPublicUserStatus(userID string) (*domain.UserStatus, error) {
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	status, err := s.repo.GetUserStatus(userID)
	if err != nil {
		return nil, err
	}

	// If user is invisible, show as offline to others
	publicStatus := status
	if status == domain.StatusInvisible {
		publicStatus = domain.StatusOffline
	}

	return &domain.UserStatus{
		UserID:       userID,
		Status:       publicStatus,
		ActualStatus: status, // Store actual status for internal use
		Timestamp:    time.Now(),
	}, nil
}

func (s *UserStatusService) GetMultipleUserStatus(userIDs []string) (map[string]*domain.UserStatus, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("user IDs cannot be empty")
	}

	statuses, err := s.repo.GetMultipleUserStatus(userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*domain.UserStatus)
	for userID, status := range statuses {
		result[userID] = &domain.UserStatus{
			UserID:    userID,
			Status:    status,
			Timestamp: time.Now(),
		}
	}

	return result, nil
}

// SendHeartbeat refreshes user online status TTL
func (s *UserStatusService) SendHeartbeat(userID string) error {
	if err := s.validateUserID(userID); err != nil {
		return err
	}

	return s.repo.RefreshUserStatusTTL(userID, domain.OnlineTTL)
}

// validateUserID validates user ID format
func (s *UserStatusService) validateUserID(userID string) error {

	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if strings.TrimSpace(userID) == "" {
		return errors.New("user ID cannot be whitespace only")
	}

	if !strings.HasPrefix(userID, "user_") {
		return errors.New("user ID must start with 'user_'")
	}

	if len(userID) > 50 {
		return errors.New("user ID too long (max 50 characters)")
	}

	return nil
}
