package services

import (
	"database/sql"
	"errors"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type ConversationService struct {
	conversationRepo    *repository.ConversationRepository
	foodPostRepo        *repository.FoodPostRepository
	hungerBroadcastRepo *repository.HungerBroadcastRepository
}

func NewConversationService(
	conversationRepo *repository.ConversationRepository,
	foodPostRepo *repository.FoodPostRepository,
	hungerBroadcastRepo *repository.HungerBroadcastRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepo:    conversationRepo,
		foodPostRepo:        foodPostRepo,
		hungerBroadcastRepo: hungerBroadcastRepo,
	}
}

func (s *ConversationService) CreateOrGetConversation(foodPostID, hungerBroadcastID *string, currentUserID, otherParticipantID string) (*models.Conversation, error) {
	if foodPostID != nil {
		// Validate food post exists
		post, err := s.foodPostRepo.FindByID(*foodPostID)
		if err != nil {
			return nil, err
		}
		if post == nil {
			return nil, errors.New("food post not found")
		}

		// Check if conversation already exists
		existingConv, err := s.conversationRepo.GetByFoodPostAndParticipants(*foodPostID, currentUserID, otherParticipantID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if existingConv != nil {
			return existingConv, nil
		}
	} else if hungerBroadcastID != nil {
		// Validate hunger broadcast exists
		broadcast, err := s.hungerBroadcastRepo.FindByID(*hungerBroadcastID)
		if err != nil {
			return nil, err
		}
		if broadcast == nil {
			return nil, errors.New("hunger broadcast not found")
		}

		// Check if conversation already exists
		existingConv, err := s.conversationRepo.GetByHungerBroadcastAndParticipants(*hungerBroadcastID, currentUserID, otherParticipantID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if existingConv != nil {
			return existingConv, nil
		}
	} else {
		// No specific context, look for a direct coordination chat
		existingConv, err := s.conversationRepo.GetByParticipantsOnly(currentUserID, otherParticipantID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if existingConv != nil {
			return existingConv, nil
		}
	}

	// Create new conversation if not exists
	return s.conversationRepo.Create(foodPostID, hungerBroadcastID, currentUserID, otherParticipantID)
}

func (s *ConversationService) GetConversation(conversationID, userID string) (*models.Conversation, error) {
	// Check if user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return nil, err
	}

	if !isParticipant {
		return nil, errors.New("unauthorized: user is not a participant in this conversation")
	}

	return s.conversationRepo.GetByID(conversationID)
}

func (s *ConversationService) GetUserConversations(userID string) ([]models.ConversationWithDetails, error) {
	return s.conversationRepo.GetUserConversations(userID)
}

func (s *ConversationService) IsParticipant(conversationID, userID string) (bool, error) {
	return s.conversationRepo.IsParticipant(conversationID, userID)
}
