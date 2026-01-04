package services

import (
	"database/sql"
	"errors"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type ConversationService struct {
	conversationRepo *repository.ConversationRepository
	foodPostRepo     *repository.FoodPostRepository
}

func NewConversationService(
	conversationRepo *repository.ConversationRepository,
	foodPostRepo *repository.FoodPostRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		foodPostRepo:     foodPostRepo,
	}
}

func (s *ConversationService) CreateOrGetConversation(foodPostID, currentUserID, otherParticipantID string) (*models.Conversation, error) {
	// Validate food post exists
	_, err := s.foodPostRepo.FindByID(foodPostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("food post not found")
		}
		return nil, err
	}

	// Check if conversation already exists
	existingConv, err := s.conversationRepo.GetByFoodPostAndParticipants(foodPostID, currentUserID, otherParticipantID)
	if err == nil {
		return existingConv, nil
	}

	// Create new conversation if not exists
	if err == sql.ErrNoRows {
		return s.conversationRepo.Create(foodPostID, currentUserID, otherParticipantID)
	}

	return nil, err
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
