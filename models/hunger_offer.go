package models

import (
	"time"
)

type HungerOffer struct {
	ID                string    `json:"id" db:"id"`
	HungerBroadcastID string    `json:"broadcastId" db:"hunger_broadcast_id"`
	UserID            string    `json:"userId" db:"user_id"`
	Message           string    `json:"message,omitempty" db:"message"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}

type CreateHungerOfferRequest struct {
	Message string `json:"message" validate:"omitempty,max=500"`
}
