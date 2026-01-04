package models

import (
	"time"
)

type Message struct {
	ID             string    `json:"id" db:"id"`
	ConversationID string    `json:"conversationId" db:"conversation_id"`
	SenderID       string    `json:"senderId" db:"sender_id"`
	Content        string    `json:"content" db:"content" validate:"required,max=2000"`
	IsRead         bool      `json:"isRead" db:"is_read"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

type MessageWithSender struct {
	Message
	SenderName string `json:"senderName" db:"sender_name"`
}

type SendMessageRequest struct {
	ConversationID string `json:"conversationId" validate:"required,uuid"`
	Content        string `json:"content" validate:"required,max=2000"`
}

// WebSocket message types
type WSMessageType string

const (
	WSMessageTypeChat            WSMessageType = "chat"
	WSMessageTypeTyping          WSMessageType = "typing"
	WSMessageTypeRead            WSMessageType = "read"
	WSMessageTypeError           WSMessageType = "error"
	WSMessageTypeConnected       WSMessageType = "connected"
	WSMessageTypeFoodPost        WSMessageType = "food_post"
	WSMessageTypeHungerBroadcast WSMessageType = "hunger_broadcast"
	WSMessageTypeRequestCreated  WSMessageType = "request_created"
	WSMessageTypeRequestUpdated  WSMessageType = "request_updated"
)

type WSMessage struct {
	Type            WSMessageType   `json:"type"`
	ConversationID  string          `json:"conversationId,omitempty"`
	Message         *Message        `json:"message,omitempty"`
	Content         string          `json:"content,omitempty"`
	Error           string          `json:"error,omitempty"`
	FoodPost        *FoodFeedItem   `json:"foodPost,omitempty"`
	HungerBroadcast *HungerFeedItem `json:"hungerBroadcast,omitempty"`
	FoodRequest     *FoodRequest    `json:"foodRequest,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
}
