package handlers

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type FoodPostHandler struct {
	foodPostService *services.FoodPostService
	requestRepo     *repository.FoodRequestRepository
}

func NewFoodPostHandler(foodPostService *services.FoodPostService, requestRepo *repository.FoodRequestRepository) *FoodPostHandler {
	return &FoodPostHandler{
		foodPostService: foodPostService,
		requestRepo:     requestRepo,
	}
}

// CreateFoodPost godoc
// @Summary Create a new food post
// @Tags food-posts
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Success 201 {object} models.FoodPost
// @Router /api/food-posts [post]
func (h *FoodPostHandler) CreateFoodPost(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.CreateFoodPostRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	// Get multiple image files if provided
	var imageFiles []*multipart.FileHeader
	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		// Try both "images" (new) and "image" (legacy) field names
		if files, ok := form.File["images"]; ok {
			imageFiles = files
		} else if files, ok := form.File["image"]; ok {
			imageFiles = files
		}
	}

	post, err := h.foodPostService.CreatePost(userID, &req, imageFiles)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, post)
}

// GetFoodPosts godoc
// @Summary Get all food posts
// @Tags food-posts
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} models.FoodPost
// @Router /api/food-posts [get]
func (h *FoodPostHandler) GetFoodPosts(c echo.Context) error {
	status := c.QueryParam("status")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 20
	}

	posts, err := h.foodPostService.GetAllPosts(status, limit, offset)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get posts")
	}

	return utils.SuccessResponse(c, http.StatusOK, posts)
}

// GetFoodPost godoc
// @Summary Get a food post by ID
// @Tags food-posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.FoodPost
// @Router /api/food-posts/{id} [get]
func (h *FoodPostHandler) GetFoodPost(c echo.Context) error {
	id := c.Param("id")

	post, err := h.foodPostService.GetPostByID(id)
	if err != nil {
		return utils.NotFound(c, "Post not found")
	}

	return utils.SuccessResponse(c, http.StatusOK, post)
}

// UpdateFoodPost godoc
// @Summary Update a food post
// @Tags food-posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Param request body models.UpdateFoodPostRequest true "Update request"
// @Success 200 {object} models.FoodPost
// @Router /api/food-posts/{id} [put]
func (h *FoodPostHandler) UpdateFoodPost(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var req models.UpdateFoodPostRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	post, err := h.foodPostService.UpdatePost(id, userID, &req)
	if err != nil {
		if err.Error() == "post not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this post" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to update post")
	}

	return utils.SuccessResponse(c, http.StatusOK, post)
}

// DeleteFoodPost godoc
// @Summary Delete a food post
// @Tags food-posts
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 204
// @Router /api/food-posts/{id} [delete]
func (h *FoodPostHandler) DeleteFoodPost(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	err := h.foodPostService.DeletePost(id, userID)
	if err != nil {
		if err.Error() == "post not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this post" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to delete post")
	}

	return c.NoContent(http.StatusNoContent)
}

// RequestFood godoc
// @Summary Request food from a post
// @Tags food-posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Param request body models.CreateFoodRequestRequest true "Request body"
// @Success 201 {object} models.FoodRequest
// @Router /api/food-posts/{id}/request [post]
func (h *FoodPostHandler) RequestFood(c echo.Context) error {
	userID := middleware.GetUserID(c)
	postID := c.Param("id")

	var req models.CreateFoodRequestRequest
	if err := c.Bind(&req); err != nil {
		// If no body, that's okay
		req.Message = ""
	}

	// Check if post exists
	_, err := h.foodPostService.GetPostByID(postID)
	if err != nil {
		return utils.NotFound(c, "Post not found")
	}

	// Check if request already exists
	exists, err := h.requestRepo.CheckExists(postID, userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to check existing request")
	}
	if exists {
		return utils.Conflict(c, "You have already requested this food")
	}

	// Create request
	foodRequest := &models.FoodRequest{
		FoodPostID: postID,
		UserID:     userID,
		Message:    req.Message,
	}

	if err := h.requestRepo.Create(foodRequest); err != nil {
		return utils.InternalServerError(c, "Failed to create request")
	}

	return utils.SuccessResponse(c, http.StatusCreated, foodRequest)
}

// GetFoodRequests godoc
// @Summary Get all requests for a food post
// @Tags food-posts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 200 {array} models.FoodRequest
// @Router /api/food-posts/{id}/requests [get]
func (h *FoodPostHandler) GetFoodRequests(c echo.Context) error {
	userID := middleware.GetUserID(c)
	postID := c.Param("id")

	// Check if user owns the post
	post, err := h.foodPostService.GetPostByID(postID)
	if err != nil {
		return utils.NotFound(c, "Post not found")
	}

	if post.UserID != userID {
		return utils.Forbidden(c, "You don't own this post")
	}

	requests, err := h.requestRepo.FindByPostID(postID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get requests")
	}

	return utils.SuccessResponse(c, http.StatusOK, requests)
}

// AcceptRequest godoc
// @Summary Accept a food request
// @Tags food-posts
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param requestId path string true "Request ID"
// @Success 200 {object} models.FoodRequest
// @Router /api/food-posts/{postId}/requests/{requestId}/accept [put]
func (h *FoodPostHandler) AcceptRequest(c echo.Context) error {
	userID := middleware.GetUserID(c)
	postID := c.Param("postId")
	requestID := c.Param("requestId")

	// Check if user owns the post
	post, err := h.foodPostService.GetPostByID(postID)
	if err != nil {
		return utils.NotFound(c, "Post not found")
	}

	if post.UserID != userID {
		return utils.Forbidden(c, "You don't own this post")
	}

	// Update request status
	if err := h.requestRepo.UpdateStatus(requestID, "accepted"); err != nil {
		return utils.InternalServerError(c, "Failed to accept request")
	}

	// Get updated request
	request, err := h.requestRepo.FindByID(requestID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get request")
	}

	return utils.SuccessResponse(c, http.StatusOK, request)
}

// RejectRequest godoc
// @Summary Reject a food request
// @Tags food-posts
// @Security BearerAuth
// @Param postId path string true "Post ID"
// @Param requestId path string true "Request ID"
// @Success 200 {object} models.FoodRequest
// @Router /api/food-posts/{postId}/requests/{requestId}/reject [put]
func (h *FoodPostHandler) RejectRequest(c echo.Context) error {
	userID := middleware.GetUserID(c)
	postID := c.Param("postId")
	requestID := c.Param("requestId")

	// Check if user owns the post
	post, err := h.foodPostService.GetPostByID(postID)
	if err != nil {
		return utils.NotFound(c, "Post not found")
	}

	if post.UserID != userID {
		return utils.Forbidden(c, "You don't own this post")
	}

	// Update request status
	if err := h.requestRepo.UpdateStatus(requestID, "rejected"); err != nil {
		return utils.InternalServerError(c, "Failed to reject request")
	}

	// Get updated request
	request, err := h.requestRepo.FindByID(requestID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get request")
	}

	return utils.SuccessResponse(c, http.StatusOK, request)
}
