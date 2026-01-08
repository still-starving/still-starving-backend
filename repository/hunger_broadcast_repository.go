package repository

import (
	"database/sql"
	"fmt"
	"strings"

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
		INSERT INTO hunger_broadcasts (id, user_id, message, location, urgency, expires_at, latitude, longitude, location_geo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ST_SetSRID(ST_MakePoint($8, $7), 4326))
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		broadcast.ID, broadcast.UserID, broadcast.Message,
		broadcast.Location, broadcast.Urgency, broadcast.ExpiresAt,
		broadcast.Latitude, broadcast.Longitude,
	).Scan(&broadcast.CreatedAt, &broadcast.UpdatedAt)
}

func (r *HungerBroadcastRepository) FindAll(status string, lat, lng, radius float64, limit, offset int) ([]models.HungerBroadcast, error) {
	var broadcasts []models.HungerBroadcast

	query := `
		SELECT hb.id, hb.user_id, hb.message, hb.location, hb.urgency, hb.status,
		       hb.expires_at, hb.created_at, hb.updated_at, u.name as user_name,
		       COALESCE(hb.latitude, 0) as latitude, COALESCE(hb.longitude, 0) as longitude
		FROM hunger_broadcasts hb
		JOIN users u ON hb.user_id = u.id
	`

	args := []interface{}{}
	argCount := 1

	whereClauses := []string{}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("hb.status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	if lat != 0 && lng != 0 && radius > 0 {
		// ST_MakePoint expects (longitude, latitude) order
		// Cast to geography for meter-based distance calculations
		whereClauses = append(whereClauses, fmt.Sprintf("ST_DWithin(hb.location_geo::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)", argCount, argCount+1, argCount+2))
		args = append(args, lng, lat, radius)
		argCount += 3
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
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
			&broadcast.Latitude, &broadcast.Longitude,
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
		       hb.expires_at, hb.created_at, hb.updated_at, u.name as user_name,
		       COALESCE(hb.latitude, 0) as latitude, COALESCE(hb.longitude, 0) as longitude
		FROM hunger_broadcasts hb
		JOIN users u ON hb.user_id = u.id
		WHERE hb.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&broadcast.ID, &broadcast.UserID, &broadcast.Message, &broadcast.Location,
		&broadcast.Urgency, &broadcast.Status, &broadcast.ExpiresAt,
		&broadcast.CreatedAt, &broadcast.UpdatedAt, &broadcast.UserName,
		&broadcast.Latitude, &broadcast.Longitude,
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
		       expires_at, created_at, updated_at, COALESCE(latitude, 0) as latitude, COALESCE(longitude, 0) as longitude
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
			&broadcast.Latitude, &broadcast.Longitude,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, broadcast)
	}

	return broadcasts, nil
}
func (r *HungerBroadcastRepository) UpdateStatus(id, status string) error {
	query := `UPDATE hunger_broadcasts SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}
func (r *HungerBroadcastRepository) FindExpiredBroadcasts() ([]models.HungerBroadcast, error) {
	var broadcasts []models.HungerBroadcast

	query := `
		SELECT id, user_id, message, location, urgency, status, expires_at, created_at, updated_at
		FROM hunger_broadcasts
		WHERE status = 'active'
		AND (
			(urgency = 'urgent' AND created_at < NOW() - INTERVAL '1 hour')
			OR
			(urgency = 'normal' AND created_at < NOW() - INTERVAL '2 hours')
		)
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.HungerBroadcast
		err := rows.Scan(
			&b.ID, &b.UserID, &b.Message, &b.Location,
			&b.Urgency, &b.Status, &b.ExpiresAt,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, b)
	}

	return broadcasts, nil
}
