package models

import (
	"time"
)

type FoodPost struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"ownerId" db:"user_id"`
	Title        string     `json:"title" db:"title" validate:"required,max=255"`
	Description  string     `json:"description" db:"description" validate:"required,max=500"`
	Quantity     string     `json:"quantity" db:"quantity" validate:"required,max=100"`
	Location     string     `json:"location" db:"location" validate:"required,max=255"`
	ExpiryDate   time.Time  `json:"expiryDate" db:"expiry_date" validate:"required"`
	ImageURLs    []string   `json:"imageUrls,omitempty"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	UserName     string     `json:"userName,omitempty" db:"user_name"`
	RequestCount int        `json:"requestCount" db:"-"` // Not from DB, populated in handler
	Price        *float64   `json:"price,omitempty" db:"price"`
	Currency     string     `json:"currency,omitempty" db:"currency"`
	SpiceLevel   string     `json:"spiceLevel" db:"spice_level"`
	Ingredients  string     `json:"ingredients" db:"ingredients"`
	CookedAt     *time.Time `json:"cookedAt,omitempty" db:"cooked_at"`
	Latitude     float64    `json:"latitude" db:"latitude"`
	Longitude    float64    `json:"longitude" db:"longitude"`
}

type CreateFoodPostRequest struct {
	Title       string     `json:"title" form:"title" validate:"required,max=255"`
	Description string     `json:"description" form:"description" validate:"required,max=500"`
	Quantity    string     `json:"quantity" form:"quantity" validate:"required,max=100"`
	Location    string     `json:"location" form:"location" validate:"required,max=255"`
	ExpiryDate  time.Time  `json:"expiryDate" form:"expiryDate" validate:"required"`
	Price       *float64   `json:"price" form:"price"`
	Currency    string     `json:"currency" form:"currency"`
	SpiceLevel  string     `json:"spiceLevel" form:"spiceLevel" validate:"omitempty,oneof=no_spicy medium_spicy spicy very_spicy"`
	Ingredients string     `json:"ingredients" form:"ingredients" validate:"omitempty,max=1000"`
	CookedAt    *time.Time `json:"cookedAt" form:"cookedAt"`
	Latitude    float64    `json:"latitude" form:"latitude" validate:"required,latitude"`
	Longitude   float64    `json:"longitude" form:"longitude" validate:"required,longitude"`
}

type UpdateFoodPostRequest struct {
	Title       string     `json:"title" validate:"omitempty,max=255"`
	Description string     `json:"description" validate:"omitempty,max=500"`
	Quantity    string     `json:"quantity" validate:"omitempty,max=100"`
	Location    string     `json:"location" validate:"omitempty,max=255"`
	ExpiryDate  time.Time  `json:"expiryDate"`
	Status      string     `json:"status" validate:"omitempty,oneof=available claimed expired"`
	Price       *float64   `json:"price" validate:"omitempty"`
	Currency    string     `json:"currency" validate:"omitempty,max=3"`
	SpiceLevel  string     `json:"spiceLevel" validate:"omitempty,oneof=no_spicy medium_spicy spicy very_spicy"`
	Ingredients string     `json:"ingredients" validate:"omitempty,max=1000"`
	CookedAt    *time.Time `json:"cookedAt" validate:"omitempty"`
	Latitude    *float64   `json:"latitude" validate:"omitempty,latitude"`
	Longitude   *float64   `json:"longitude" validate:"omitempty,longitude"`
}

type FoodPostWithOwnership struct {
	FoodPost
	IsOwner bool `json:"isOwner"`
}

// For feed response
type FoodFeedItem struct {
	Type        string     `json:"type"`
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Quantity    string     `json:"quantity"`
	Location    string     `json:"location"`
	ExpiryDate  time.Time  `json:"expiryDate"`
	Status      string     `json:"status"`
	OwnerName   string     `json:"ownerName"`
	OwnerID     string     `json:"ownerId"`
	ImageURLs   []string   `json:"imageUrls,omitempty"`
	IsOwner     bool       `json:"isOwner"`
	Price       *float64   `json:"price,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	SpiceLevel  string     `json:"spiceLevel"`
	Ingredients string     `json:"ingredients"`
	CookedAt    *time.Time `json:"cookedAt,omitempty"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
}
