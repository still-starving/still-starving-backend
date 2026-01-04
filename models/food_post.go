package models

import (
	"time"
)

type FoodPost struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"ownerId" db:"user_id"`
	Title       string    `json:"title" db:"title" validate:"required,max=255"`
	Description string    `json:"description" db:"description" validate:"required,max=500"`
	Quantity    string    `json:"quantity" db:"quantity" validate:"required,max=100"`
	Location    string    `json:"location" db:"location" validate:"required,max=255"`
	ExpiryDate  time.Time `json:"expiryDate" db:"expiry_date" validate:"required"`
	ImageURLs   []string  `json:"imageUrls,omitempty"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	UserName    string    `json:"userName,omitempty" db:"user_name"`
}

type CreateFoodPostRequest struct {
	Title       string    `json:"title" form:"title" validate:"required,max=255"`
	Description string    `json:"description" form:"description" validate:"required,max=500"`
	Quantity    string    `json:"quantity" form:"quantity" validate:"required,max=100"`
	Location    string    `json:"location" form:"location" validate:"required,max=255"`
	ExpiryDate  time.Time `json:"expiryDate" form:"expiryDate" validate:"required"`
}

type UpdateFoodPostRequest struct {
	Title       string    `json:"title" validate:"omitempty,max=255"`
	Description string    `json:"description" validate:"omitempty,max=500"`
	Quantity    string    `json:"quantity" validate:"omitempty,max=100"`
	Location    string    `json:"location" validate:"omitempty,max=255"`
	ExpiryDate  time.Time `json:"expiryDate"`
	Status      string    `json:"status" validate:"omitempty,oneof=available claimed expired"`
}

type FoodPostWithOwnership struct {
	FoodPost
	IsOwner bool `json:"isOwner"`
}

// For feed response
type FoodFeedItem struct {
	Type        string    `json:"type"`
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Quantity    string    `json:"quantity"`
	Location    string    `json:"location"`
	ExpiryDate  time.Time `json:"expiryDate"`
	Status      string    `json:"status"`
	OwnerName   string    `json:"ownerName"`
	OwnerID     string    `json:"ownerId"`
	ImageURLs   []string  `json:"imageUrls,omitempty"`
	IsOwner     bool      `json:"isOwner"`
}
