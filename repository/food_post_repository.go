package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type FoodPostRepository struct {
	db *sql.DB
}

func NewFoodPostRepository(db *sql.DB) *FoodPostRepository {
	return &FoodPostRepository{db: db}
}

func (r *FoodPostRepository) Create(post *models.FoodPost) error {
	post.ID = uuid.New().String()
	post.Status = "available"

	query := `
		INSERT INTO food_posts (id, user_id, title, description, quantity, location, expiry_date, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		post.ID, post.UserID, post.Title, post.Description,
		post.Quantity, post.Location, post.ExpiryDate, post.ImageURL,
	).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (r *FoodPostRepository) FindAll(status string, limit, offset int) ([]models.FoodPost, error) {
	var posts []models.FoodPost

	query := `
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.image_url, fp.status, fp.created_at, fp.updated_at,
		       u.name as user_name
		FROM food_posts fp
		JOIN users u ON fp.user_id = u.id
	`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE fp.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY fp.created_at DESC"

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
		var post models.FoodPost
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Description,
			&post.Quantity, &post.Location, &post.ExpiryDate, &post.ImageURL,
			&post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UserName,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *FoodPostRepository) FindByID(id string) (*models.FoodPost, error) {
	post := &models.FoodPost{}

	query := `
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.image_url, fp.status, fp.created_at, fp.updated_at,
		       u.name as user_name
		FROM food_posts fp
		JOIN users u ON fp.user_id = u.id
		WHERE fp.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Description,
		&post.Quantity, &post.Location, &post.ExpiryDate, &post.ImageURL,
		&post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UserName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return post, err
}

func (r *FoodPostRepository) Update(post *models.FoodPost) error {
	query := `
		UPDATE food_posts
		SET title = $1, description = $2, quantity = $3, location = $4,
		    expiry_date = $5, status = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING updated_at
	`

	return r.db.QueryRow(
		query,
		post.Title, post.Description, post.Quantity, post.Location,
		post.ExpiryDate, post.Status, post.ID,
	).Scan(&post.UpdatedAt)
}

func (r *FoodPostRepository) Delete(id string) error {
	query := `DELETE FROM food_posts WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *FoodPostRepository) FindByUserID(userID string) ([]models.FoodPost, error) {
	var posts []models.FoodPost

	query := `
		SELECT id, user_id, title, description, quantity, location,
		       expiry_date, image_url, status, created_at, updated_at
		FROM food_posts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.FoodPost
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Description,
			&post.Quantity, &post.Location, &post.ExpiryDate, &post.ImageURL,
			&post.Status, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}
