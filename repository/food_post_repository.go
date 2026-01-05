package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		INSERT INTO food_posts (id, user_id, title, description, quantity, location, expiry_date, price, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		post.ID, post.UserID, post.Title, post.Description,
		post.Quantity, post.Location, post.ExpiryDate,
		post.Price, post.Currency,
	).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (r *FoodPostRepository) FindAll(status string, limit, offset int) ([]models.FoodPost, error) {
	var posts []models.FoodPost

	query := `
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.status, fp.created_at, fp.updated_at,
		       u.name as user_name, fp.price, fp.currency,
		       COALESCE(array_agg(pi.image_url ORDER BY pi.display_order) FILTER (WHERE pi.image_url IS NOT NULL), '{}') as image_urls
		FROM food_posts fp
		JOIN users u ON fp.user_id = u.id
		LEFT JOIN post_images pi ON fp.id = pi.food_post_id
	`

	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE fp.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " GROUP BY fp.id, u.name"
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
		var imageURLs pq.StringArray
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Description,
			&post.Quantity, &post.Location, &post.ExpiryDate,
			&post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UserName,
			&post.Price, &post.Currency,
			&imageURLs,
		)
		if err != nil {
			return nil, err
		}
		post.ImageURLs = []string(imageURLs)
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *FoodPostRepository) FindByID(id string) (*models.FoodPost, error) {
	post := &models.FoodPost{}

	query := `
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.status, fp.created_at, fp.updated_at,
		       u.name as user_name, fp.price, fp.currency,
		       COALESCE(array_agg(pi.image_url ORDER BY pi.display_order) FILTER (WHERE pi.image_url IS NOT NULL), '{}') as image_urls
		FROM food_posts fp
		JOIN users u ON fp.user_id = u.id
		LEFT JOIN post_images pi ON fp.id = pi.food_post_id
		WHERE fp.id = $1
		GROUP BY fp.id, u.name
	`

	var imageURLs pq.StringArray
	err := r.db.QueryRow(query, id).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Description,
		&post.Quantity, &post.Location, &post.ExpiryDate,
		&post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UserName,
		&post.Price, &post.Currency,
		&imageURLs,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	post.ImageURLs = []string(imageURLs)
	return post, nil
}

func (r *FoodPostRepository) Update(post *models.FoodPost) error {
	query := `
		UPDATE food_posts
		SET title = $1, description = $2, quantity = $3, location = $4,
		    expiry_date = $5, status = $6, price = $7, currency = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at
	`

	return r.db.QueryRow(
		query,
		post.Title, post.Description, post.Quantity, post.Location,
		post.ExpiryDate, post.Status, post.Price, post.Currency, post.ID,
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
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.status, fp.created_at, fp.updated_at, fp.price, fp.currency,
		       COALESCE(array_agg(pi.image_url ORDER BY pi.display_order) FILTER (WHERE pi.image_url IS NOT NULL), '{}') as image_urls
		FROM food_posts fp
		LEFT JOIN post_images pi ON fp.id = pi.food_post_id
		WHERE fp.user_id = $1
		GROUP BY fp.id
		ORDER BY fp.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.FoodPost
		var imageURLs pq.StringArray
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Description,
			&post.Quantity, &post.Location, &post.ExpiryDate,
			&post.Status, &post.CreatedAt, &post.UpdatedAt,
			&post.Price, &post.Currency,
			&imageURLs,
		)
		if err != nil {
			return nil, err
		}
		post.ImageURLs = []string(imageURLs)
		posts = append(posts, post)
	}

	return posts, nil
}
