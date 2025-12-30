package models

import (
	"time"
)

type HungerBroadcast struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	Message   string    `json:"message" db:"message" validate:"required,max=140"`
	Location  string    `json:"location" db:"location" validate:"required,max=255"`
	Urgency   string    `json:"urgency" db:"urgency" validate:"oneof=normal urgent"`
	Status    string    `json:"status" db:"status"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	UserName  string    `json:"userName,omitempty" db:"user_name"`
}

type CreateHungerBroadcastRequest struct {
	Message  string `json:"message" validate:"required,max=140"`
	Location string `json:"location" validate:"required,max=255"`
	Urgency  string `json:"urgency" validate:"required,oneof=normal urgent"`
}

type HungerBroadcastWithOwnership struct {
	HungerBroadcast
	IsOwner bool `json:"isOwner"`
}

// For feed response
type HungerFeedItem struct {
	Type       string    `json:"type"`
	ID         string    `json:"id"`
	Message    string    `json:"message"`
	Location   string    `json:"location"`
	Urgency    string    `json:"urgency"`
	UserName   string    `json:"userName"`
	TimePosted time.Time `json:"timePosted"`
	IsOwner    bool      `json:"isOwner"`
}
