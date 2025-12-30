package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	user.ID = uuid.New().String()

	query := `
		INSERT INTO users (id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(query, user.ID, user.Name, user.Email, user.PasswordHash).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`

	return r.db.QueryRow(query, user.Name, user.Email, user.ID).Scan(&user.UpdatedAt)
}

func (r *UserRepository) GetStats(userID string) (map[string]int, error) {
	stats := make(map[string]int)

	// Count food shared
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM food_posts WHERE user_id = $1
	`, userID).Scan(&stats["foodShared"])
	if err != nil {
		return nil, err
	}

	// Count people helped (accepted requests)
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM food_requests fr
		JOIN food_posts fp ON fr.food_post_id = fp.id
		WHERE fp.user_id = $1 AND fr.status = 'accepted'
	`, userID).Scan(&stats["peopleHelped"])
	if err != nil {
		return nil, err
	}

	// Count active posts
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM food_posts
		WHERE user_id = $1 AND status = 'available'
	`, userID).Scan(&stats["activePosts"])
	if err != nil {
		return nil, err
	}

	return stats, nil
}
