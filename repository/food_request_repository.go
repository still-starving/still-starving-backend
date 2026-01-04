package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type FoodRequestRepository struct {
	db *sql.DB
}

func NewFoodRequestRepository(db *sql.DB) *FoodRequestRepository {
	return &FoodRequestRepository{db: db}
}

func (r *FoodRequestRepository) Create(request *models.FoodRequest) error {
	request.ID = uuid.New().String()
	request.Status = "pending"

	query := `
		INSERT INTO food_requests (id, food_post_id, user_id, message)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		request.ID, request.FoodPostID, request.UserID, request.Message,
	).Scan(&request.CreatedAt, &request.UpdatedAt)
}

func (r *FoodRequestRepository) FindByPostID(postID string) ([]models.FoodRequest, error) {
	var requests []models.FoodRequest

	query := `
		SELECT fr.id, fr.food_post_id, fr.user_id, fr.message, fr.status,
		       fr.created_at, fr.updated_at, u.name as user_name
		FROM food_requests fr
		JOIN users u ON fr.user_id = u.id
		WHERE fr.food_post_id = $1
		ORDER BY fr.created_at DESC
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.FoodRequest
		err := rows.Scan(
			&request.ID, &request.FoodPostID, &request.UserID, &request.Message,
			&request.Status, &request.CreatedAt, &request.UpdatedAt, &request.UserName,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (r *FoodRequestRepository) FindByUserID(userID string) ([]models.FoodRequestWithPostInfo, error) {
	var requests []models.FoodRequestWithPostInfo

	query := `
		SELECT fr.id, fr.food_post_id, fr.user_id, fr.message, fr.status,
		       fr.created_at, fr.updated_at,
		       fp.title as post_title, u.name as post_owner_name, fp.user_id as post_owner_id
		FROM food_requests fr
		JOIN food_posts fp ON fr.food_post_id = fp.id
		JOIN users u ON fp.user_id = u.id
		WHERE fr.user_id = $1
		ORDER BY fr.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.FoodRequestWithPostInfo
		err := rows.Scan(
			&request.ID, &request.FoodPostID, &request.UserID, &request.Message,
			&request.Status, &request.CreatedAt, &request.UpdatedAt,
			&request.PostTitle, &request.PostOwnerName, &request.PostOwnerID,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (r *FoodRequestRepository) FindByID(id string) (*models.FoodRequest, error) {
	request := &models.FoodRequest{}

	query := `
		SELECT id, food_post_id, user_id, message, status, created_at, updated_at
		FROM food_requests
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&request.ID, &request.FoodPostID, &request.UserID, &request.Message,
		&request.Status, &request.CreatedAt, &request.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return request, err
}

func (r *FoodRequestRepository) UpdateStatus(id, status string) error {
	query := `
		UPDATE food_requests
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *FoodRequestRepository) CheckExists(postID, userID string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM food_requests
			WHERE food_post_id = $1 AND user_id = $2
		)
	`

	err := r.db.QueryRow(query, postID, userID).Scan(&exists)
	return exists, err
}
