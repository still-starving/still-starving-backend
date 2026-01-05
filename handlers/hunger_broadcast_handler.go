package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/middleware"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/repository"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type HungerBroadcastHandler struct {
	hungerBroadcastService *services.HungerBroadcastService
	offerRepo              *repository.HungerOfferRepository
	hub                    *services.Hub
}

func NewHungerBroadcastHandler(hungerBroadcastService *services.HungerBroadcastService, offerRepo *repository.HungerOfferRepository, hub *services.Hub) *HungerBroadcastHandler {
	return &HungerBroadcastHandler{
		hungerBroadcastService: hungerBroadcastService,
		offerRepo:              offerRepo,
		hub:                    hub,
	}
}

// CreateHungerBroadcast godoc
// @Summary Create a hunger broadcast
// @Tags hunger-broadcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateHungerBroadcastRequest true "Broadcast request"
// @Success 201 {object} models.HungerBroadcast
// @Router /api/hunger-broadcasts [post]
func (h *HungerBroadcastHandler) CreateHungerBroadcast(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.CreateHungerBroadcastRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	broadcast, err := h.hungerBroadcastService.CreateBroadcast(userID, &req)
	if err != nil {
		return utils.InternalServerError(c, "Failed to create broadcast")
	}

	// Broadcast the new item to all connected clients
	feedItem := &models.HungerFeedItem{
		Type:       "hunger",
		ID:         broadcast.ID,
		Message:    broadcast.Message,
		Location:   broadcast.Location,
		Urgency:    broadcast.Urgency,
		UserName:   broadcast.UserName, // Note: UserName might be empty here
		OwnerID:    broadcast.UserID,
		TimePosted: broadcast.CreatedAt,
		IsOwner:    false, // For broadcast, receiver is not owner
	}

	wsMsg := models.WSMessage{
		Type:            models.WSMessageTypeHungerBroadcast,
		HungerBroadcast: feedItem,
		Timestamp:       time.Now(),
	}

	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.hub.BroadcastToAll(msgBytes)
	}

	return utils.SuccessResponse(c, http.StatusCreated, broadcast)
}

// GetHungerBroadcasts godoc
// @Summary Get all hunger broadcasts
// @Tags hunger-broadcasts
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} models.HungerBroadcast
// @Router /api/hunger-broadcasts [get]
func (h *HungerBroadcastHandler) GetHungerBroadcasts(c echo.Context) error {
	status := c.QueryParam("status")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 20
	}

	broadcasts, err := h.hungerBroadcastService.GetAllBroadcasts(status, limit, offset)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get broadcasts")
	}

	return utils.SuccessResponse(c, http.StatusOK, broadcasts)
}

// GetHungerBroadcast godoc
// @Summary Get a hunger broadcast by ID
// @Tags hunger-broadcasts
// @Produce json
// @Param id path string true "Broadcast ID"
// @Success 200 {object} models.HungerBroadcast
// @Router /api/hunger-broadcasts/{id} [get]
func (h *HungerBroadcastHandler) GetHungerBroadcast(c echo.Context) error {
	id := c.Param("id")

	broadcast, err := h.hungerBroadcastService.GetBroadcastByID(id)
	if err != nil {
		return utils.NotFound(c, "Broadcast not found")
	}

	return utils.SuccessResponse(c, http.StatusOK, broadcast)
}

// DeleteHungerBroadcast godoc
// @Summary Delete a hunger broadcast
// @Tags hunger-broadcasts
// @Security BearerAuth
// @Param id path string true "Broadcast ID"
// @Success 204
// @Router /api/hunger-broadcasts/{id} [delete]
func (h *HungerBroadcastHandler) DeleteHungerBroadcast(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	err := h.hungerBroadcastService.DeleteBroadcast(id, userID)
	if err != nil {
		if err.Error() == "broadcast not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this broadcast" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to delete broadcast")
	}

	return c.NoContent(http.StatusNoContent)
}

// OfferFood godoc
// @Summary Offer food to a hunger broadcast
// @Tags hunger-broadcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Broadcast ID"
// @Param request body models.CreateHungerOfferRequest true "Offer request"
// @Success 201 {object} models.HungerOffer
// @Router /api/hunger-broadcasts/{id}/offer [post]
func (h *HungerBroadcastHandler) OfferFood(c echo.Context) error {
	userID := middleware.GetUserID(c)
	broadcastID := c.Param("id")

	var req models.CreateHungerOfferRequest
	if err := c.Bind(&req); err != nil {
		// If no body, that's okay
		req.Message = ""
	}

	// Check if broadcast exists
	_, err := h.hungerBroadcastService.GetBroadcastByID(broadcastID)
	if err != nil {
		return utils.NotFound(c, "Broadcast not found")
	}

	// Check if offer already exists
	exists, err := h.offerRepo.CheckExists(broadcastID, userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to check existing offer")
	}
	if exists {
		return utils.Conflict(c, "You have already offered food to this broadcast")
	}

	// Create offer
	offer := &models.HungerOffer{
		HungerBroadcastID: broadcastID,
		UserID:            userID,
		Message:           req.Message,
	}

	if err := h.offerRepo.Create(offer); err != nil {
		return utils.InternalServerError(c, "Failed to create offer")
	}

	return utils.SuccessResponse(c, http.StatusCreated, offer)
}

// ResolveHungerBroadcast godoc
// @Summary Resolve a hunger broadcast
// @Tags hunger-broadcasts
// @Security BearerAuth
// @Param id path string true "Broadcast ID"
// @Success 200
// @Router /api/hunger-broadcasts/{id}/resolve [put]
func (h *HungerBroadcastHandler) ResolveHungerBroadcast(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	err := h.hungerBroadcastService.ResolveBroadcast(id, userID)
	if err != nil {
		if err.Error() == "broadcast not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this broadcast" {
			return utils.Forbidden(c, err.Error())
		}
		log.Printf("Error resolving broadcast: %v", err)
		return utils.ErrorResponseJSON(c, 500, "INTERNAL_SERVER_ERROR", "Failed to resolve broadcast", err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]string{"message": "Broadcast marked as resolved"})
}
