package services

import (
	"fmt"
	"mime/multipart"

	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
)

type FoodPostService struct {
	foodPostRepo  *repository.FoodPostRepository
	postImageRepo *repository.PostImageRepository
	imageService  *ImageService
}

func NewFoodPostService(foodPostRepo *repository.FoodPostRepository, postImageRepo *repository.PostImageRepository, imageService *ImageService) *FoodPostService {
	return &FoodPostService{
		foodPostRepo:  foodPostRepo,
		postImageRepo: postImageRepo,
		imageService:  imageService,
	}
}

func (s *FoodPostService) CreatePost(userID string, req *models.CreateFoodPostRequest, imageFiles []*multipart.FileHeader) (*models.FoodPost, error) {
	post := &models.FoodPost{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Quantity:    req.Quantity,
		Location:    req.Location,
		ExpiryDate:  req.ExpiryDate,
		Price:       req.Price,
		Currency:    req.Currency,
		SpiceLevel:  req.SpiceLevel,
		Ingredients: req.Ingredients,
		CookedAt:    req.CookedAt,
		Packaging:   req.Packaging,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	}

	// Create post in database first
	if err := s.foodPostRepo.Create(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Upload images if provided
	var uploadedURLs []string
	if len(imageFiles) > 0 {
		for i, file := range imageFiles {
			imageURL, err := s.imageService.UploadImage(file)
			if err != nil {
				// Rollback: delete uploaded images and the post
				for _, url := range uploadedURLs {
					_ = s.imageService.DeleteImage(url)
				}
				_ = s.foodPostRepo.Delete(post.ID)
				return nil, fmt.Errorf("failed to upload image: %w", err)
			}
			uploadedURLs = append(uploadedURLs, imageURL)

			// Save image record to database
			_, err = s.postImageRepo.Create(post.ID, imageURL, i)
			if err != nil {
				// Rollback: delete uploaded images and the post
				for _, url := range uploadedURLs {
					_ = s.imageService.DeleteImage(url)
				}
				_ = s.foodPostRepo.Delete(post.ID)
				return nil, fmt.Errorf("failed to save image record: %w", err)
			}
		}
		post.ImageURLs = uploadedURLs
	}

	return post, nil
}

func (s *FoodPostService) GetAllPosts(status string, limit, offset int) ([]models.FoodPost, error) {
	return s.foodPostRepo.FindAll(status, 0, 0, 0, limit, offset)
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
	if req.Price != nil {
		post.Price = req.Price
	}
	if req.Currency != "" {
		post.Currency = req.Currency
	}
	if req.SpiceLevel != "" {
		post.SpiceLevel = req.SpiceLevel
	}
	if req.Ingredients != "" {
		post.Ingredients = req.Ingredients
	}
	if req.CookedAt != nil {
		post.CookedAt = req.CookedAt
	}
	if req.Latitude != nil {
		post.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		post.Longitude = *req.Longitude
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

	// Delete images from MinIO
	for _, url := range post.ImageURLs {
		_ = s.imageService.DeleteImage(url)
	}

	// Delete from database (cascade will delete post_images records)
	if err := s.foodPostRepo.Delete(postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (s *FoodPostService) GetUserPosts(userID string) ([]models.FoodPost, error) {
	return s.foodPostRepo.FindByUserID(userID)
}

// AddImages adds new images to an existing post
func (s *FoodPostService) AddImages(postID, userID string, imageFiles []*multipart.FileHeader) (*models.FoodPost, error) {
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

	// Get current max display order
	startOrder := len(post.ImageURLs)

	// Upload and save new images
	for i, file := range imageFiles {
		imageURL, err := s.imageService.UploadImage(file)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}

		_, err = s.postImageRepo.Create(postID, imageURL, startOrder+i)
		if err != nil {
			_ = s.imageService.DeleteImage(imageURL)
			return nil, fmt.Errorf("failed to save image record: %w", err)
		}

		post.ImageURLs = append(post.ImageURLs, imageURL)
	}

	return post, nil
}

func (s *FoodPostService) MarkAsClaimed(postID, userID string) error {
	return s.foodPostRepo.UpdateClaimedInfo(postID, userID)
}

func (s *FoodPostService) SubmitFeedback(postID, userID string, req *models.FoodPostFeedbackRequest) error {
	// Get post to verify claim
	post, err := s.foodPostRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return fmt.Errorf("post not found")
	}

	// Verify user is the one who claimed the post
	if post.ClaimedByUserID == nil || *post.ClaimedByUserID != userID {
		return fmt.Errorf("unauthorized: only the person who claimed this food can leave feedback")
	}

	// Verify if feedback was already given
	if post.Rating != nil {
		return fmt.Errorf("feedback has already been submitted for this post")
	}

	return s.foodPostRepo.UpdateFeedback(postID, req.Rating, req.Review)
}

func (s *FoodPostService) AutoCloseExpiredPosts() error {
	expiredPosts, err := s.foodPostRepo.FindExpiredAvailablePosts()
	if err != nil {
		return err
	}

	for _, post := range expiredPosts {
		// Update status to expired
		if err := s.foodPostRepo.UpdateStatus(post.ID, "expired"); err != nil {
			fmt.Printf("Failed to expire post %s: %v\n", post.ID, err)
			continue
		}
		// TODO: Option to notify user via WebSocket that their post has expired
	}

	return nil
}
