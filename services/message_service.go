package services

import (
	"errors"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type MessageService struct {
	messageRepo      *repository.MessageRepository
	conversationRepo *repository.ConversationRepository
}

func NewMessageService(
	messageRepo *repository.MessageRepository,
	conversationRepo *repository.ConversationRepository,
) *MessageService {
	return &MessageService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
	}
}

func (s *MessageService) SendMessage(conversationID, senderID, content string) (*models.Message, error) {
	// Validate sender is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, senderID)
	if err != nil {
		return nil, err
	}

	if !isParticipant {
		return nil, errors.New("unauthorized: user is not a participant in this conversation")
	}

	// Create message
	message, err := s.messageRepo.Create(conversationID, senderID, content)
	if err != nil {
		return nil, err
	}

	// Update conversation's last message time
	err = s.conversationRepo.UpdateLastMessageTime(conversationID, message.CreatedAt)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetMessages(conversationID, userID string, limit, offset int) ([]models.MessageWithSender, error) {
	// Validate user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return nil, err
	}

	if !isParticipant {
		return nil, errors.New("unauthorized: user is not a participant in this conversation")
	}

	return s.messageRepo.GetByConversationID(conversationID, limit, offset)
}

func (s *MessageService) MarkAsRead(conversationID, userID string) error {
	// Validate user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return err
	}

	if !isParticipant {
		return errors.New("unauthorized: user is not a participant in this conversation")
	}

	return s.messageRepo.MarkAsRead(conversationID, userID)
}

func (s *MessageService) GetUnreadCount(userID string) (int, error) {
	return s.messageRepo.GetUnreadCount(userID)
}
