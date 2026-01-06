package services

import (
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type FeedService struct {
	foodPostRepo        *repository.FoodPostRepository
	hungerBroadcastRepo *repository.HungerBroadcastRepository
	userRepo            *repository.UserRepository
}

func NewFeedService(
	foodPostRepo *repository.FoodPostRepository,
	hungerBroadcastRepo *repository.HungerBroadcastRepository,
	userRepo *repository.UserRepository,
) *FeedService {
	return &FeedService{
		foodPostRepo:        foodPostRepo,
		hungerBroadcastRepo: hungerBroadcastRepo,
		userRepo:            userRepo,
	}
}

func (s *FeedService) GetFeed(feedType, userID string, lat, lng, radius float64) ([]interface{}, error) {
	// If radius is not specified (0) and location is provided
	if radius <= 0 && lat != 0 && lng != 0 {
		// Use user's preference if authenticated
		if userID != "" {
			user, err := s.userRepo.FindByID(userID)
			if err == nil && user != nil {
				radius = user.PreferredRadiusKm * 1000 // Convert km to meters
			}
		}

		// Fallback to 1km if still not set
		if radius <= 0 {
			radius = 1000 // 1km default
		}
	}

	var feed []interface{}

	// Get food posts if requested
	if feedType == "" || feedType == "all" || feedType == "food" {
		foodPosts, err := s.foodPostRepo.FindAll("available", lat, lng, radius, 20, 0)
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
				OwnerID:     post.UserID,
				ImageURLs:   post.ImageURLs,
				IsOwner:     post.UserID == userID,
				Price:       post.Price,
				Currency:    post.Currency,
				SpiceLevel:  post.SpiceLevel,
				Ingredients: post.Ingredients,
				CookedAt:    post.CookedAt,
				Latitude:    post.Latitude,
				Longitude:   post.Longitude,
			}
			feed = append(feed, feedItem)
		}
	}

	// Get hunger broadcasts if requested
	if feedType == "" || feedType == "all" || feedType == "hunger" {
		broadcasts, err := s.hungerBroadcastRepo.FindAll("active", lat, lng, radius, 20, 0)
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
				OwnerID:    broadcast.UserID,
				TimePosted: broadcast.CreatedAt,
				IsOwner:    broadcast.UserID == userID,
				Latitude:   broadcast.Latitude,
				Longitude:  broadcast.Longitude,
			}
			feed = append(feed, feedItem)
		}
	}

	return feed, nil
}
