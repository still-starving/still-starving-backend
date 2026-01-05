package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

type MessageHandler struct {
	messageService *services.MessageService
	imageService   *services.ImageService
}

func NewMessageHandler(messageService *services.MessageService, imageService *services.ImageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		imageService:   imageService,
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

func (h *MessageHandler) UploadMessageImage(c echo.Context) error {
	file, err := c.FormFile("image")
	if err != nil {
		return utils.BadRequest(c, "Image is required", nil)
	}

	url, err := h.imageService.UploadImage(file)
	if err != nil {
		return utils.InternalServerError(c, "Failed to upload image")
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]string{
		"imageUrl": url,
	})
}

func (h *MessageHandler) SendMessage(c echo.Context) error {
	userID := c.Get("userID").(string)
	conversationID := c.Param("id")

	var req models.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	if err := c.Validate(&req); err != nil {
		return utils.BadRequest(c, "Validation failed", utils.FormatValidationErrors(err))
	}

	message, err := h.messageService.SendMessage(conversationID, userID, req.Content, req.Type, req.Metadata)
	if err != nil {
		if err.Error() == "unauthorized: user is not a participant in this conversation" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, "Failed to send message")
	}

	return utils.SuccessResponse(c, http.StatusCreated, message)
}
