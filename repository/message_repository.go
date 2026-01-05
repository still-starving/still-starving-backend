package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(conversationID, senderID, content, msgType string, metadata map[string]interface{}) (*models.Message, error) {
	message := &models.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		Type:           msgType,
		Metadata:       metadata,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content, type, metadata, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, conversation_id, sender_id, content, type, metadata, is_read, created_at
	`

	err = r.db.QueryRow(
		query,
		message.ID,
		message.ConversationID,
		message.SenderID,
		message.Content,
		message.Type,
		metadataJSON,
		message.IsRead,
		message.CreatedAt,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.Content,
		&message.Type,
		&metadataJSON,
		&message.IsRead,
		&message.CreatedAt,
	)

	if err == nil && len(metadataJSON) > 0 && string(metadataJSON) != "null" {
		json.Unmarshal(metadataJSON, &message.Metadata)
	}

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *MessageRepository) GetByConversationID(conversationID string, limit, offset int) ([]models.MessageWithSender, error) {
	query := `
		SELECT 
			m.id,
			m.conversation_id,
			m.sender_id,
			m.content,
			m.type,
			m.metadata,
			m.is_read,
			m.created_at,
			u.name as sender_name
		FROM messages m
		INNER JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.MessageWithSender
	var metadataJSON []byte
	for rows.Next() {
		var msg models.MessageWithSender
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Content,
			&msg.Type,
			&metadataJSON,
			&msg.IsRead,
			&msg.CreatedAt,
			&msg.SenderName,
		)
		if err != nil {
			return nil, err
		}
		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			json.Unmarshal(metadataJSON, &msg.Metadata)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MessageRepository) MarkAsRead(conversationID, userID string) error {
	query := `
		UPDATE messages
		SET is_read = TRUE
		WHERE conversation_id = $1
		AND sender_id != $2
		AND is_read = FALSE
	`

	_, err := r.db.Exec(query, conversationID, userID)
	return err
}

func (r *MessageRepository) GetUnreadCount(userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM messages m
		INNER JOIN conversations c ON m.conversation_id = c.id
		WHERE (c.participant_1_id = $1 OR c.participant_2_id = $1)
		AND m.sender_id != $1
		AND m.is_read = FALSE
	`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *MessageRepository) GetUnreadCountByConversation(conversationID, userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE conversation_id = $1
		AND sender_id != $2
		AND is_read = FALSE
	`

	err := r.db.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
