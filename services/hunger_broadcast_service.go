package services

import (
	"fmt"
	"time"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type HungerBroadcastService struct {
	hungerBroadcastRepo *repository.HungerBroadcastRepository
}

func NewHungerBroadcastService(hungerBroadcastRepo *repository.HungerBroadcastRepository) *HungerBroadcastService {
	return &HungerBroadcastService{
		hungerBroadcastRepo: hungerBroadcastRepo,
	}
}

func (s *HungerBroadcastService) CreateBroadcast(userID string, req *models.CreateHungerBroadcastRequest) (*models.HungerBroadcast, error) {
	// Calculate expiry time (12 hours from now)
	expiresAt := time.Now().Add(12 * time.Hour)

	broadcast := &models.HungerBroadcast{
		UserID:    userID,
		Message:   req.Message,
		Location:  req.Location,
		Urgency:   req.Urgency,
		ExpiresAt: expiresAt,
	}

	if err := s.hungerBroadcastRepo.Create(broadcast); err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	return broadcast, nil
}

func (s *HungerBroadcastService) GetAllBroadcasts(status string, limit, offset int) ([]models.HungerBroadcast, error) {
	return s.hungerBroadcastRepo.FindAll(status, limit, offset)
}

func (s *HungerBroadcastService) GetBroadcastByID(id string) (*models.HungerBroadcast, error) {
	broadcast, err := s.hungerBroadcastRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast not found")
	}
	return broadcast, nil
}

func (s *HungerBroadcastService) DeleteBroadcast(broadcastID, userID string) error {
	// Get existing broadcast
	broadcast, err := s.hungerBroadcastRepo.FindByID(broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast: %w", err)
	}
	if broadcast == nil {
		return fmt.Errorf("broadcast not found")
	}

	// Check ownership
	if broadcast.UserID != userID {
		return fmt.Errorf("unauthorized: you don't own this broadcast")
	}

	// Delete from database
	if err := s.hungerBroadcastRepo.Delete(broadcastID); err != nil {
		return fmt.Errorf("failed to delete broadcast: %w", err)
	}

	return nil
}

func (s *HungerBroadcastService) GetUserBroadcasts(userID string) ([]models.HungerBroadcast, error) {
	return s.hungerBroadcastRepo.FindByUserID(userID)
}
