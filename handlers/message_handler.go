package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

func (h *MessageHandler) GetMessages(c echo.Context) error {
	userID := c.Get("userID").(string)
	conversationID := c.Param("id")

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, err := h.messageService.GetMessages(conversationID, userID, limit, offset)
	if err != nil {
		if err.Error() == "unauthorized: user is not a participant in this conversation" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to retrieve messages")
	}

	return utils.SuccessResponse(c, http.StatusOK, messages)
}

func (h *MessageHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get("userID").(string)
	conversationID := c.Param("id")

	err := h.messageService.MarkAsRead(conversationID, userID)
	if err != nil {
		if err.Error() == "unauthorized: user is not a participant in this conversation" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to mark messages as read")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]string{"message": "Messages marked as read"})
}

func (h *MessageHandler) GetUnreadCount(c echo.Context) error {
	userID := c.Get("userID").(string)

	count, err := h.messageService.GetUnreadCount(userID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to get unread count")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]int{
		"unreadCount": count,
	})
}
