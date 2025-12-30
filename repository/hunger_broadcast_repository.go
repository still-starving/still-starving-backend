package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type HungerBroadcastRepository struct {
	db *sql.DB
}

func NewHungerBroadcastRepository(db *sql.DB) *HungerBroadcastRepository {
	return &HungerBroadcastRepository{db: db}
}

func (r *HungerBroadcastRepository) Create(broadcast *models.HungerBroadcast) error {
	broadcast.ID = uuid.New().String()
	broadcast.Status = "active"

	query := `
		INSERT INTO hunger_broadcasts (id, user_id, message, location, urgency, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		broadcast.ID, broadcast.UserID, broadcast.Message,
		broadcast.Location, broadcast.Urgency, broadcast.ExpiresAt,
	).Scan(&broadcast.CreatedAt, &broadcast.UpdatedAt)
}

func (r *HungerBroadcastRepository) FindAll(status string, limit, offset int) ([]models.HungerBroadcast, error) {
	var broadcasts []models.HungerBroadcast

	query := `
		SELECT hb.id, hb.user_id, hb.message, hb.location, hb.urgency, hb.status,
		       hb.expires_at, hb.created_at, hb.updated_at, u.name as user_name
		FROM hunger_broadcasts hb
		JOIN users u ON hb.user_id = u.id
	`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE hb.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY hb.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var broadcast models.HungerBroadcast
		err := rows.Scan(
			&broadcast.ID, &broadcast.UserID, &broadcast.Message, &broadcast.Location,
			&broadcast.Urgency, &broadcast.Status, &broadcast.ExpiresAt,
			&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.UserName,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, broadcast)
	}

	return broadcasts, nil
}

func (r *HungerBroadcastRepository) FindByID(id string) (*models.HungerBroadcast, error) {
	broadcast := &models.HungerBroadcast{}

	query := `
		SELECT hb.id, hb.user_id, hb.message, hb.location, hb.urgency, hb.status,
		       hb.expires_at, hb.created_at, hb.updated_at, u.name as user_name
		FROM hunger_broadcasts hb
		JOIN users u ON hb.user_id = u.id
		WHERE hb.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&broadcast.ID, &broadcast.UserID, &broadcast.Message, &broadcast.Location,
		&broadcast.Urgency, &broadcast.Status, &broadcast.ExpiresAt,
		&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.UserName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return broadcast, err
}

func (r *HungerBroadcastRepository) Delete(id string) error {
	query := `DELETE FROM hunger_broadcasts WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *HungerBroadcastRepository) FindByUserID(userID string) ([]models.HungerBroadcast, error) {
	var broadcasts []models.HungerBroadcast

	query := `
		SELECT id, user_id, message, location, urgency, status,
		       expires_at, created_at, updated_at
		FROM hunger_broadcasts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var broadcast models.HungerBroadcast
		err := rows.Scan(
			&broadcast.ID, &broadcast.UserID, &broadcast.Message, &broadcast.Location,
			&broadcast.Urgency, &broadcast.Status, &broadcast.ExpiresAt,
			&broadcast.CreatedAt, &broadcast.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, broadcast)
	}

	return broadcasts, nil
}
