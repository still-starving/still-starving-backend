package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type HungerBroadcastService struct {
	hungerBroadcastRepo *repository.HungerBroadcastRepository
	hungerOfferRepo     *repository.HungerOfferRepository
	notificationRepo    *repository.NotificationRepository
	hub                 *Hub
}

func NewHungerBroadcastService(
	hungerBroadcastRepo *repository.HungerBroadcastRepository,
	hungerOfferRepo *repository.HungerOfferRepository,
	notificationRepo *repository.NotificationRepository,
	hub *Hub,
) *HungerBroadcastService {
	return &HungerBroadcastService{
		hungerBroadcastRepo: hungerBroadcastRepo,
		hungerOfferRepo:     hungerOfferRepo,
		notificationRepo:    notificationRepo,
		hub:                 hub,
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
func (s *HungerBroadcastService) ResolveBroadcast(broadcastID, userID string) error {
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

	// Update status to fulfilled
	if err := s.hungerBroadcastRepo.UpdateStatus(broadcastID, "fulfilled"); err != nil {
		return fmt.Errorf("failed to fulfill broadcast: %w", err)
	}

	// Broadcast global fulfillment to everyone to remove from feed
	globalWsMsg := models.WSMessage{
		Type:      models.WSMessageTypeHungerBroadcastExpired,
		Content:   broadcastID,
		Timestamp: time.Now(),
	}
	if msgBytes, err := json.Marshal(globalWsMsg); err == nil {
		s.hub.BroadcastToAll(msgBytes)
	}

	return nil
}
func (s *HungerBroadcastService) AutoCloseExpiredBroadcasts() error {
	expired, err := s.hungerBroadcastRepo.FindExpiredBroadcasts()
	if err != nil {
		return err
	}

	for _, b := range expired {
		// Update status to expired
		if err := s.hungerBroadcastRepo.UpdateStatus(b.ID, "expired"); err != nil {
			continue
		}

		// Check if any offers were made
		offers, err := s.hungerOfferRepo.FindByBroadcastID(b.ID)
		if err == nil && len(offers) == 0 {
			// Send optimistic notification
			notif := &models.Notification{
				UserID:      b.UserID,
				Title:       "Hunger Broadcast Concluded",
				Message:     "Your community post has concluded. Though no one could join this time, we're so glad you reached out! Every voice strengthens our community, and we hope to see you post again soon.",
				Type:        models.NotificationTypeBroadcastClosed,
				ReferenceID: &b.ID,
			}
			if err := s.notificationRepo.Create(notif); err == nil {
				// Broadcast via WebSocket to the user (Private)
				wsMsg := models.WSMessage{
					Type:         models.WSMessageTypeNotification,
					Notification: notif,
					Timestamp:    time.Now(),
				}
				if msgBytes, err := json.Marshal(wsMsg); err == nil {
					s.hub.BroadcastToUsers([]string{b.UserID}, msgBytes)
				}
			}
		}

		// Broadcast global expiration to everyone to remove from feed (Public)
		globalWsMsg := models.WSMessage{
			Type:      models.WSMessageTypeHungerBroadcastExpired,
			Content:   b.ID, // Send ID so frontend knows which one to remove
			Timestamp: time.Now(),
		}
		if msgBytes, err := json.Marshal(globalWsMsg); err == nil {
			s.hub.BroadcastToAll(msgBytes)
		}
	}

	return nil
}
