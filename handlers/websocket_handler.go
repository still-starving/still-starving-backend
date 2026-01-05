package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/yourusername/food-sharing-backend/models"
	"github.com/yourusername/food-sharing-backend/services"
	"github.com/yourusername/food-sharing-backend/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, validate origin properly
		return true
	},
}

type WebSocketHandler struct {
	hub                 *services.Hub
	messageService      *services.MessageService
	conversationService *services.ConversationService
	jwtSecret           string
}

func NewWebSocketHandler(
	hub *services.Hub,
	messageService *services.MessageService,
	conversationService *services.ConversationService,
	jwtSecret string,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:                 hub,
		messageService:      messageService,
		conversationService: conversationService,
		jwtSecret:           jwtSecret,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	// Get token from query parameter
	tokenString := c.QueryParam("token")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Missing authentication token",
		})
	}

	// Validate JWT token
	claims, err := utils.ValidateToken(tokenString, h.jwtSecret)
	if err != nil {
		log.Printf("WebSocket JWT validation failed - Error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired authentication token",
		})
	}

	userID := claims.UserID
	log.Printf("WebSocket connection authenticated for user: %s", userID)

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return err
	}

	// Create client
	client := &services.Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register client
	client.Hub.Register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump(h.handleMessage)

	return nil
}

func (h *WebSocketHandler) handleMessage(client *services.Client, message []byte) {
	var wsMsg models.WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		h.sendError(client, "Invalid message format")
		return
	}

	switch wsMsg.Type {
	case models.WSMessageTypeChat:
		h.handleChatMessage(client, &wsMsg)
	case models.WSMessageTypeRead:
		h.handleReadMessage(client, &wsMsg)
	case models.WSMessageTypeTyping:
		h.handleTypingMessage(client, &wsMsg)
	default:
		h.sendError(client, "Unknown message type")
	}
}

func (h *WebSocketHandler) handleChatMessage(client *services.Client, wsMsg *models.WSMessage) {
	if wsMsg.ConversationID == "" || wsMsg.Content == "" {
		h.sendError(client, "Missing conversation ID or content")
		return
	}

	// Validate user is participant
	isParticipant, err := h.conversationService.IsParticipant(wsMsg.ConversationID, client.UserID)
	if err != nil || !isParticipant {
		h.sendError(client, "Unauthorized: not a participant in this conversation")
		return
	}

	// Save message to database
	// Passing wsMsg.Message.Type if it exists, otherwise mapping WSMessage.Type if appropriate
	actualMsgType := "text"
	if wsMsg.Message != nil && wsMsg.Message.Type != "" {
		actualMsgType = wsMsg.Message.Type
	} else if string(wsMsg.Type) == "image" || string(wsMsg.Type) == "price_offer" {
		actualMsgType = string(wsMsg.Type)
	}

	savedMessage, err := h.messageService.SendMessage(wsMsg.ConversationID, client.UserID, wsMsg.Content, actualMsgType, wsMsg.Metadata)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		h.sendError(client, "Failed to send message")
		return
	}

	// Get conversation to find other participant
	conversation, err := h.conversationService.GetConversation(wsMsg.ConversationID, client.UserID)
	if err != nil {
		log.Printf("Failed to get conversation: %v", err)
		return
	}

	// Determine other participant
	otherParticipantID := conversation.Participant1ID
	if otherParticipantID == client.UserID {
		otherParticipantID = conversation.Participant2ID
	}

	// Broadcast to both participants using "chat" event type
	responseMsg := models.WSMessage{
		Type:           models.WSMessageTypeChat,
		ConversationID: wsMsg.ConversationID,
		Message:        savedMessage,
		Metadata:       savedMessage.Metadata,
		Timestamp:      time.Now(),
	}

	msgBytes, err := json.Marshal(responseMsg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	// Send to both sender and receiver
	h.hub.BroadcastToUsers([]string{client.UserID, otherParticipantID}, msgBytes)
}

func (h *WebSocketHandler) handleReadMessage(client *services.Client, wsMsg *models.WSMessage) {
	if wsMsg.ConversationID == "" {
		h.sendError(client, "Missing conversation ID")
		return
	}

	err := h.messageService.MarkAsRead(wsMsg.ConversationID, client.UserID)
	if err != nil {
		log.Printf("Failed to mark messages as read: %v", err)
		h.sendError(client, "Failed to mark messages as read")
		return
	}

	// Get conversation to find other participant
	conversation, err := h.conversationService.GetConversation(wsMsg.ConversationID, client.UserID)
	if err != nil {
		log.Printf("Failed to get conversation: %v", err)
		return
	}

	// Determine other participant
	otherParticipantID := conversation.Participant1ID
	if otherParticipantID == client.UserID {
		otherParticipantID = conversation.Participant2ID
	}

	// Notify other participant
	responseMsg := models.WSMessage{
		Type:           models.WSMessageTypeRead,
		ConversationID: wsMsg.ConversationID,
		Timestamp:      time.Now(),
	}

	msgBytes, err := json.Marshal(responseMsg)
	if err != nil {
		log.Printf("Failed to marshal read message: %v", err)
		return
	}

	h.hub.BroadcastToUsers([]string{otherParticipantID}, msgBytes)
}

func (h *WebSocketHandler) handleTypingMessage(client *services.Client, wsMsg *models.WSMessage) {
	if wsMsg.ConversationID == "" {
		return
	}

	// Get conversation to find other participant
	conversation, err := h.conversationService.GetConversation(wsMsg.ConversationID, client.UserID)
	if err != nil {
		return
	}

	// Determine other participant
	otherParticipantID := conversation.Participant1ID
	if otherParticipantID == client.UserID {
		otherParticipantID = conversation.Participant2ID
	}

	// Forward typing indicator to other participant
	responseMsg := models.WSMessage{
		Type:           models.WSMessageTypeTyping,
		ConversationID: wsMsg.ConversationID,
		Timestamp:      time.Now(),
	}

	msgBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return
	}

	h.hub.BroadcastToUsers([]string{otherParticipantID}, msgBytes)
}

func (h *WebSocketHandler) sendError(client *services.Client, errorMsg string) {
	errResponse := models.WSMessage{
		Type:      models.WSMessageTypeError,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(errResponse)
	if err != nil {
		log.Printf("Failed to marshal error message: %v", err)
		return
	}

	select {
	case client.Send <- msgBytes:
	default:
		log.Printf("Failed to send error message to client")
	}
}
