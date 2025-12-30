package models

import (
	"time"
)

type FoodRequest struct {
	ID         string    `json:"id" db:"id"`
	FoodPostID string    `json:"postId" db:"food_post_id"`
	UserID     string    `json:"userId" db:"user_id"`
	Message    string    `json:"message,omitempty" db:"message"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
	UserName   string    `json:"userName,omitempty" db:"user_name"`
}

type CreateFoodRequestRequest struct {
	Message string `json:"message" validate:"omitempty,max=500"`
}

type FoodRequestWithPostInfo struct {
	FoodRequest
	PostTitle     string `json:"postTitle" db:"post_title"`
	PostOwnerName string `json:"postOwnerName" db:"post_owner_name"`
}
