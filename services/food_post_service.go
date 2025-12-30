package services

import (
	"fmt"
	"mime/multipart"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type FoodPostService struct {
	foodPostRepo *repository.FoodPostRepository
	imageService *ImageService
}

func NewFoodPostService(foodPostRepo *repository.FoodPostRepository, imageService *ImageService) *FoodPostService {
	return &FoodPostService{
		foodPostRepo: foodPostRepo,
		imageService: imageService,
	}
}

func (s *FoodPostService) CreatePost(userID string, req *models.CreateFoodPostRequest, imageFile *multipart.FileHeader) (*models.FoodPost, error) {
	post := &models.FoodPost{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Quantity:    req.Quantity,
		Location:    req.Location,
		ExpiryDate:  req.ExpiryDate,
	}

	// Upload image if provided
	if imageFile != nil {
		imageURL, err := s.imageService.UploadImage(imageFile)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}
		post.ImageURL = imageURL
	}

	// Create post in database
	if err := s.foodPostRepo.Create(post); err != nil {
		// If database creation fails and image was uploaded, delete the image
		if post.ImageURL != "" {
			_ = s.imageService.DeleteImage(post.ImageURL)
		}
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (s *FoodPostService) GetAllPosts(status string, limit, offset int) ([]models.FoodPost, error) {
	return s.foodPostRepo.FindAll(status, limit, offset)
}

func (s *FoodPostService) GetPostByID(id string) (*models.FoodPost, error) {
	post, err := s.foodPostRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (s *FoodPostService) UpdatePost(postID, userID string, req *models.UpdateFoodPostRequest) (*models.FoodPost, error) {
	// Get existing post
	post, err := s.foodPostRepo.FindByID(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}

	// Check ownership
	if post.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this post")
	}

	// Update fields
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Description != "" {
		post.Description = req.Description
	}
	if req.Quantity != "" {
		post.Quantity = req.Quantity
	}
	if req.Location != "" {
		post.Location = req.Location
	}
	if !req.ExpiryDate.IsZero() {
		post.ExpiryDate = req.ExpiryDate
	}
	if req.Status != "" {
		post.Status = req.Status
	}

	// Update in database
	if err := s.foodPostRepo.Update(post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

func (s *FoodPostService) DeletePost(postID, userID string) error {
	// Get existing post
	post, err := s.foodPostRepo.FindByID(postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return fmt.Errorf("post not found")
	}

	// Check ownership
	if post.UserID != userID {
		return fmt.Errorf("unauthorized: you don't own this post")
	}

	// Delete image if exists
	if post.ImageURL != "" {
		_ = s.imageService.DeleteImage(post.ImageURL)
	}

	// Delete from database
	if err := s.foodPostRepo.Delete(postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (s *FoodPostService) GetUserPosts(userID string) ([]models.FoodPost, error) {
	return s.foodPostRepo.FindByUserID(userID)
}
