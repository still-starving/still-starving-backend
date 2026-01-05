package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	requests := make([]models.FoodRequest, 0)

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
	requests := make([]models.FoodRequestWithPostInfo, 0)

	query := `
		SELECT fr.id, fr.food_post_id, fr.user_id, fr.message, fr.status,
		       fr.created_at, fr.updated_at, fr.viewed_at,
		       fp.title as post_title, u.name as post_owner_name, fp.user_id as post_owner_id,
		       c.id as conversation_id
		FROM food_requests fr
		JOIN food_posts fp ON fr.food_post_id = fp.id
		JOIN users u ON fp.user_id = u.id
		LEFT JOIN conversations c ON c.food_post_id = fp.id 
		  AND (c.participant_1_id = fr.user_id OR c.participant_2_id = fr.user_id)
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
			&request.Status, &request.CreatedAt, &request.UpdatedAt, &request.ViewedAt,
			&request.PostTitle, &request.PostOwnerName, &request.PostOwnerID,
			&request.ConversationID,
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
			WHERE food_post_id = $1 AND user_id = $2 AND status = 'pending'
		)
	`

	err := r.db.QueryRow(query, postID, userID).Scan(&exists)
	return exists, err
}

func (r *FoodRequestRepository) CountPendingByOwnerID(ownerID string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM food_requests fr
		JOIN food_posts fp ON fr.food_post_id = fp.id
		WHERE fp.user_id = $1 AND fr.status = 'pending'
	`

	err := r.db.QueryRow(query, ownerID).Scan(&count)
	return count, err
}

func (r *FoodRequestRepository) GetPendingCountsByPostIDs(postIDs []string) (map[string]int, error) {
	counts := make(map[string]int)

	if len(postIDs) == 0 {
		return counts, nil
	}

	query := `
		SELECT food_post_id, COUNT(*) as count
		FROM food_requests
		WHERE food_post_id = ANY($1) AND status = 'pending'
		GROUP BY food_post_id
	`

	rows, err := r.db.Query(query, pq.Array(postIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID string
		var count int
		if err := rows.Scan(&postID, &count); err != nil {
			return nil, err
		}
		counts[postID] = count
	}

	return counts, nil
}

func (r *FoodRequestRepository) CountUnviewedByUserID(userID string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM food_requests
		WHERE user_id = $1 
		  AND (status = 'accepted' OR status = 'rejected')
		  AND viewed_at IS NULL
	`

	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *FoodRequestRepository) MarkAsViewedByUserID(userID string) error {
	query := `
		UPDATE food_requests
		SET viewed_at = NOW()
		WHERE user_id = $1
	`

	_, err := r.db.Exec(query, userID)
	return err
}
