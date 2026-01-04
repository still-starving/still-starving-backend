package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(foodPostID, participant1ID, participant2ID string) (*models.Conversation, error) {
	conversation := &models.Conversation{
		ID:             uuid.New().String(),
		FoodPostID:     foodPostID,
		Participant1ID: participant1ID,
		Participant2ID: participant2ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	query := `
		INSERT INTO conversations (id, food_post_id, participant_1_id, participant_2_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, food_post_id, participant_1_id, participant_2_id, last_message_at, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		conversation.ID,
		conversation.FoodPostID,
		conversation.Participant1ID,
		conversation.Participant2ID,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(
		&conversation.ID,
		&conversation.FoodPostID,
		&conversation.Participant1ID,
		&conversation.Participant2ID,
		&conversation.LastMessageAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) GetByID(id string) (*models.Conversation, error) {
	conversation := &models.Conversation{}
	query := `
		SELECT id, food_post_id, participant_1_id, participant_2_id, last_message_at, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&conversation.ID,
		&conversation.FoodPostID,
		&conversation.Participant1ID,
		&conversation.Participant2ID,
		&conversation.LastMessageAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) GetByFoodPostAndParticipants(foodPostID, participant1ID, participant2ID string) (*models.Conversation, error) {
	conversation := &models.Conversation{}
	query := `
		SELECT id, food_post_id, participant_1_id, participant_2_id, last_message_at, created_at, updated_at
		FROM conversations
		WHERE food_post_id = $1
		AND (
			(participant_1_id = $2 AND participant_2_id = $3)
			OR
			(participant_1_id = $3 AND participant_2_id = $2)
		)
	`

	err := r.db.QueryRow(query, foodPostID, participant1ID, participant2ID).Scan(
		&conversation.ID,
		&conversation.FoodPostID,
		&conversation.Participant1ID,
		&conversation.Participant2ID,
		&conversation.LastMessageAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) GetUserConversations(userID string) ([]models.ConversationWithDetails, error) {
	query := `
		SELECT 
			c.id,
			c.food_post_id,
			c.participant_1_id,
			c.participant_2_id,
			c.last_message_at,
			c.created_at,
			c.updated_at,
			fp.title as food_post_title,
			CASE 
				WHEN c.participant_1_id = $1 THEN c.participant_2_id
				ELSE c.participant_1_id
			END as other_participant_id,
			CASE 
				WHEN c.participant_1_id = $1 THEN u2.name
				ELSE u1.name
			END as other_participant_name,
			(
				SELECT content 
				FROM messages 
				WHERE conversation_id = c.id 
				ORDER BY created_at DESC 
				LIMIT 1
			) as last_message_content,
			(
				SELECT COUNT(*) 
				FROM messages 
				WHERE conversation_id = c.id 
				AND sender_id != $1 
				AND is_read = FALSE
			) as unread_count
		FROM conversations c
		INNER JOIN food_posts fp ON c.food_post_id = fp.id
		INNER JOIN users u1 ON c.participant_1_id = u1.id
		INNER JOIN users u2 ON c.participant_2_id = u2.id
		WHERE c.participant_1_id = $1 OR c.participant_2_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.ConversationWithDetails
	for rows.Next() {
		var conv models.ConversationWithDetails
		err := rows.Scan(
			&conv.ID,
			&conv.FoodPostID,
			&conv.Participant1ID,
			&conv.Participant2ID,
			&conv.LastMessageAt,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&conv.FoodPostTitle,
			&conv.OtherParticipantID,
			&conv.OtherParticipantName,
			&conv.LastMessageContent,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *ConversationRepository) UpdateLastMessageTime(conversationID string, timestamp time.Time) error {
	query := `
		UPDATE conversations
		SET last_message_at = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, timestamp, time.Now(), conversationID)
	return err
}

func (r *ConversationRepository) IsParticipant(conversationID, userID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM conversations
		WHERE id = $1 AND (participant_1_id = $2 OR participant_2_id = $2)
	`

	err := r.db.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
