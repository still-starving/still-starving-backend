package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourusername/food-sharing-backend/models"
)

type Client struct {
	ID        string
	UserID    string
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	closeOnce sync.Once
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.Send)
		c.Conn.Close()
	})
}

type Hub struct {
	// Registered clients mapped by user ID
	Clients map[string]*Client

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast messages to specific users
	Broadcast chan *BroadcastMessage

	// Broadcast messages to all connected clients
	BroadcastAll chan []byte

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

type BroadcastMessage struct {
	UserIDs []string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[string]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan *BroadcastMessage),
		BroadcastAll: make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// Disconnect existing client with same user ID if exists
			if existingClient, exists := h.Clients[client.UserID]; exists {
				existingClient.Close()
				delete(h.Clients, client.UserID)
				log.Printf("Disconnected existing client for user %s", client.UserID)
			}
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client registered: UserID=%s, Total clients=%d", client.UserID, len(h.Clients))

			// Send connected confirmation
			connectedMsg := models.WSMessage{
				Type:      models.WSMessageTypeConnected,
				Timestamp: time.Now(),
			}
			if msgBytes, err := json.Marshal(connectedMsg); err == nil {
				select {
				case client.Send <- msgBytes:
					log.Printf("Sent connected confirmation to user %s", client.UserID)
				default:
					log.Printf("Warning: Could not send connected message to user %s (channel full)", client.UserID)
				}
			}

		case client := <-h.Unregister:
			h.mu.Lock()
			// Only delete if this is the current active client for this user
			if activeClient, ok := h.Clients[client.UserID]; ok && activeClient.ID == client.ID {
				delete(h.Clients, client.UserID)
				client.Close()
				log.Printf("Client unregistered: UserID=%s, Total clients=%d", client.UserID, len(h.Clients))
			} else {
				// Even if it's not the active one, ensure it's closed
				client.Close()
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, userID := range message.UserIDs {
				if client, ok := h.Clients[userID]; ok {
					select {
					case client.Send <- message.Message:
					default:
						// If send fails, it's likely a dead connection.
						// We don't close it here because Unregister or ReadPump will handle it.
						// Closing here while Hub is running could cause confusion.
						log.Printf("Warning: Failed to send message to user %s (channel full)", userID)
					}
				}
			}
			h.mu.RUnlock()

		case message := <-h.BroadcastAll:
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					log.Printf("Warning: Failed to broadcast to user %s (channel full)", client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToAll(message []byte) {
	h.BroadcastAll <- message
}

func (h *Hub) BroadcastToUsers(userIDs []string, message []byte) {
	h.Broadcast <- &BroadcastMessage{
		UserIDs: userIDs,
		Message: message,
	}
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.Clients[userID]
	return exists
}

// Client read pump
func (c *Client) ReadPump(messageHandler func(*Client, []byte)) {
	defer func() {
		c.Hub.Unregister <- c
		c.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		messageHandler(c, message)
	}
}

// Client write pump
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
