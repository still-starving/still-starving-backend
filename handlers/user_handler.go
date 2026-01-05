package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type UserHandler struct {
	userRepo               *repository.UserRepository
	foodRequestRepo        *repository.FoodRequestRepository
	foodPostService        *services.FoodPostService
	hungerBroadcastService *services.HungerBroadcastService
}

func NewUserHandler(
	userRepo *repository.UserRepository,
	foodRequestRepo *repository.FoodRequestRepository,
	foodPostService *services.FoodPostService,
	hungerBroadcastService *services.HungerBroadcastService,
) *UserHandler {
	return &UserHandler{
		userRepo:               userRepo,
		foodRequestRepo:        foodRequestRepo,
		foodPostService:        foodPostService,
		hungerBroadcastService: hungerBroadcastService,
	}
}

// GetMyRequests godoc
// @Summary Get all requests made by current user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.FoodRequestWithPostInfo
// @Router /api/my-requests [get]
func (h *UserHandler) GetMyRequests(c echo.Context) error {
	userID := middleware.GetUserID(c)

	requests, err := h.foodRequestRepo.FindByUserID(userID)
	if err != nil {
		log.Printf("ERROR in GetMyRequests: %v", err)
		return utils.InternalServerError(c, "Failed to get requests")
	}

	return utils.SuccessResponse(c, http.StatusOK, requests)
}

// GetPendingRequestsCount godoc
// @Summary Get count of pending requests for current user's posts
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/my-food-requests/pending-count [get]
func (h *UserHandler) GetPendingRequestsCount(c echo.Context) error {
	userID := middleware.GetUserID(c)

	count, err := h.foodRequestRepo.CountPendingByOwnerID(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get pending requests count")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// GetMyRequestsUnviewedCount godoc
// @Summary Get count of unviewed request updates for current user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/my-requests/unviewed-count [get]
func (h *UserHandler) GetMyRequestsUnviewedCount(c echo.Context) error {
	userID := middleware.GetUserID(c)

	count, err := h.foodRequestRepo.CountUnviewedByUserID(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get unviewed count")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// MarkMyRequestsAsViewed godoc
// @Summary Mark all user's requests as viewed
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/my-requests/mark-viewed [put]
func (h *UserHandler) MarkMyRequestsAsViewed(c echo.Context) error {
	userID := middleware.GetUserID(c)

	err := h.foodRequestRepo.MarkAsViewedByUserID(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to mark requests as viewed")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GetMyPosts godoc
// @Summary Get all posts by current user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/my-posts [get]
func (h *UserHandler) GetMyPosts(c echo.Context) error {
	userID := middleware.GetUserID(c)

	foodPosts, err := h.foodPostService.GetUserPosts(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get food posts")
	}

	hungerBroadcasts, err := h.hungerBroadcastService.GetUserBroadcasts(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get hunger broadcasts")
	}

	// Get pending request counts for all food posts
	postIDs := make([]string, len(foodPosts))
	for i, post := range foodPosts {
		postIDs[i] = post.ID
	}

	requestCounts, err := h.foodRequestRepo.GetPendingCountsByPostIDs(postIDs)
	if err != nil {
		// Log error but don't fail - just set counts to 0
		requestCounts = make(map[string]int)
	}

	// Populate request counts
	for i := range foodPosts {
		foodPosts[i].RequestCount = requestCounts[foodPosts[i].ID]
	}

	response := map[string]interface{}{
		"foodPosts":        foodPosts,
		"hungerBroadcasts": hungerBroadcasts,
	}

	return utils.SuccessResponse(c, http.StatusOK, response)
}

// GetMyHungerBroadcasts godoc
// @Summary Get all hunger broadcasts by current user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.HungerBroadcast
// @Router /api/my-hunger-broadcasts [get]
func (h *UserHandler) GetMyHungerBroadcasts(c echo.Context) error {
	userID := middleware.GetUserID(c)

	broadcasts, err := h.hungerBroadcastService.GetUserBroadcasts(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get broadcasts")
	}

	return utils.SuccessResponse(c, http.StatusOK, broadcasts)
}

// GetProfile godoc
// @Summary Get user profile with statistics
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/profile [get]
func (h *UserHandler) GetProfile(c echo.Context) error {
	userID := middleware.GetUserID(c)

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.NotFound(c, "User not found")
	}

	stats, err := h.userRepo.GetStats(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get statistics")
	}

	response := map[string]interface{}{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"createdAt": user.CreatedAt,
		"stats":     stats,
	}

	return utils.SuccessResponse(c, http.StatusOK, response)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Update request"
// @Success 200 {object} models.User
// @Router /api/profile [put]
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Name  string `json:"name" validate:"omitempty,min=2,max=255"`
		Email string `json:"email" validate:"omitempty,email"`
	}

	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.NotFound(c, "User not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := h.userRepo.Update(user); err != nil {
		return utils.InternalServerError(c, "Failed to update profile")
	}

	return utils.SuccessResponse(c, http.StatusOK, user)
}
