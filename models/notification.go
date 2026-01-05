package models

import (
	"time"
)

type Notification struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"userId" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Message     string    `json:"message" db:"message"`
	Type        string    `json:"type" db:"type"`
	ReferenceID *string   `json:"referenceId,omitempty" db:"reference_id"`
	IsRead      bool      `json:"isRead" db:"is_read"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

const (
	NotificationTypeBroadcastClosed = "broadcast_closed"
)
