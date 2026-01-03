package models

import (
	"time"
)

type PostImage struct {
	ID           string    `json:"id" db:"id"`
	FoodPostID   string    `json:"foodPostId" db:"food_post_id"`
	ImageURL     string    `json:"imageUrl" db:"image_url"`
	DisplayOrder int       `json:"displayOrder" db:"display_order"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
