package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type ConversationHandler struct {
	conversationService *services.ConversationService
}

func NewConversationHandler(conversationService *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

func (h *ConversationHandler) CreateOrGetConversation(c echo.Context) error {
	userID := c.Get("userID").(string)

	var req models.CreateConversationRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, err.Error(), nil)
	}

	var foodPostID *string
	if req.FoodPostID != "" {
		foodPostID = &req.FoodPostID
	}

	var hungerBroadcastID *string
	if req.HungerBroadcastID != "" {
		hungerBroadcastID = &req.HungerBroadcastID
	}

	conversation, err := h.conversationService.CreateOrGetConversation(foodPostID, hungerBroadcastID, userID, req.OtherParticipantID)
	if err != nil {
		if err.Error() == "food post not found" || err.Error() == "hunger broadcast not found" {
			return utils.NotFound(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to create conversation")
	}

	return utils.SuccessResponse(c, http.StatusOK, conversation)
}

func (h *ConversationHandler) GetUserConversations(c echo.Context) error {
	userID := c.Get("userID").(string)

	conversations, err := h.conversationService.GetUserConversations(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to retrieve conversations")
	}

	return utils.SuccessResponse(c, http.StatusOK, conversations)
}

func (h *ConversationHandler) GetConversation(c echo.Context) error {
	userID := c.Get("userID").(string)
	conversationID := c.Param("id")

	conversation, err := h.conversationService.GetConversation(conversationID, userID)
	if err != nil {
		if err.Error() == "unauthorized: user is not a participant in this conversation" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to retrieve conversation")
	}

	return utils.SuccessResponse(c, http.StatusOK, conversation)
}
