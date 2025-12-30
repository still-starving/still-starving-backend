package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type HungerOfferRepository struct {
	db *sql.DB
}

func NewHungerOfferRepository(db *sql.DB) *HungerOfferRepository {
	return &HungerOfferRepository{db: db}
}

func (r *HungerOfferRepository) Create(offer *models.HungerOffer) error {
	offer.ID = uuid.New().String()

	query := `
		INSERT INTO hunger_offers (id, hunger_broadcast_id, user_id, message)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	return r.db.QueryRow(
		query,
		offer.ID, offer.HungerBroadcastID, offer.UserID, offer.Message,
	).Scan(&offer.CreatedAt)
}

func (r *HungerOfferRepository) FindByBroadcastID(broadcastID string) ([]models.HungerOffer, error) {
	var offers []models.HungerOffer

	query := `
		SELECT id, hunger_broadcast_id, user_id, message, created_at
		FROM hunger_offers
		WHERE hunger_broadcast_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, broadcastID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var offer models.HungerOffer
		err := rows.Scan(
			&offer.ID, &offer.HungerBroadcastID, &offer.UserID,
			&offer.Message, &offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

func (r *HungerOfferRepository) CheckExists(broadcastID, userID string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM hunger_offers
			WHERE hunger_broadcast_id = $1 AND user_id = $2
		)
	`

	err := r.db.QueryRow(query, broadcastID, userID).Scan(&exists)
	return exists, err
}
