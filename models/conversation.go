package models

import (
	"time"
)

type Conversation struct {
	ID             string     `json:"id" db:"id"`
	FoodPostID     string     `json:"foodPostId" db:"food_post_id"`
	Participant1ID string     `json:"participant1Id" db:"participant_1_id"`
	Participant2ID string     `json:"participant2Id" db:"participant_2_id"`
	LastMessageAt  *time.Time `json:"lastMessageAt,omitempty" db:"last_message_at"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

type ConversationWithDetails struct {
	Conversation
	FoodPostTitle        string  `json:"foodPostTitle" db:"food_post_title"`
	OtherParticipantID   string  `json:"otherParticipantId" db:"other_participant_id"`
	OtherParticipantName string  `json:"otherParticipantName" db:"other_participant_name"`
	LastMessageContent   *string `json:"lastMessageContent,omitempty" db:"last_message_content"`
	UnreadCount          int     `json:"unreadCount" db:"unread_count"`
}

type CreateConversationRequest struct {
	FoodPostID         string `json:"foodPostId" validate:"required,uuid"`
	OtherParticipantID string `json:"otherParticipantId" validate:"required,uuid"`
}
