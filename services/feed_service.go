package services

import (
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type FeedService struct {
	foodPostRepo        *repository.FoodPostRepository
	hungerBroadcastRepo *repository.HungerBroadcastRepository
}

func NewFeedService(
	foodPostRepo *repository.FoodPostRepository,
	hungerBroadcastRepo *repository.HungerBroadcastRepository,
) *FeedService {
	return &FeedService{
		foodPostRepo:        foodPostRepo,
		hungerBroadcastRepo: hungerBroadcastRepo,
	}
}

func (s *FeedService) GetFeed(feedType, userID string) ([]interface{}, error) {
	var feed []interface{}

	// Get food posts if requested
	if feedType == "" || feedType == "all" || feedType == "food" {
		foodPosts, err := s.foodPostRepo.FindAll("available", 20, 0)
		if err != nil {
			return nil, err
		}

		for _, post := range foodPosts {
			feedItem := models.FoodFeedItem{
				Type:        "food",
				ID:          post.ID,
				Title:       post.Title,
				Description: post.Description,
				Quantity:    post.Quantity,
				Location:    post.Location,
				ExpiryDate:  post.ExpiryDate,
				Status:      post.Status,
				OwnerName:   post.UserName,
				ImageURLs:   post.ImageURLs,
				IsOwner:     post.UserID == userID,
			}
			feed = append(feed, feedItem)
		}
	}

	// Get hunger broadcasts if requested
	if feedType == "" || feedType == "all" || feedType == "hunger" {
		broadcasts, err := s.hungerBroadcastRepo.FindAll("active", 20, 0)
		if err != nil {
			return nil, err
		}

		for _, broadcast := range broadcasts {
			feedItem := models.HungerFeedItem{
				Type:       "hunger",
				ID:         broadcast.ID,
				Message:    broadcast.Message,
				Location:   broadcast.Location,
				Urgency:    broadcast.Urgency,
				UserName:   broadcast.UserName,
				TimePosted: broadcast.CreatedAt,
				IsOwner:    broadcast.UserID == userID,
			}
			feed = append(feed, feedItem)
		}
	}

	return feed, nil
}
