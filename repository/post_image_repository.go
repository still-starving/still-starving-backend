package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type PostImageRepository struct {
	db *sql.DB
}

func NewPostImageRepository(db *sql.DB) *PostImageRepository {
	return &PostImageRepository{db: db}
}

func (r *PostImageRepository) Create(postID string, imageURL string, displayOrder int) (*models.PostImage, error) {
	image := &models.PostImage{
		ID:           uuid.New().String(),
		FoodPostID:   postID,
		ImageURL:     imageURL,
		DisplayOrder: displayOrder,
	}

	query := `
		INSERT INTO post_images (id, food_post_id, image_url, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.db.QueryRow(query, image.ID, image.FoodPostID, image.ImageURL, image.DisplayOrder).Scan(&image.CreatedAt)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (r *PostImageRepository) CreateBatch(postID string, imageURLs []string) error {
	for i, url := range imageURLs {
		_, err := r.Create(postID, url, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostImageRepository) FindByPostID(postID string) ([]models.PostImage, error) {
	var images []models.PostImage

	query := `
		SELECT id, food_post_id, image_url, display_order, created_at
		FROM post_images
		WHERE food_post_id = $1
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var image models.PostImage
		err := rows.Scan(&image.ID, &image.FoodPostID, &image.ImageURL, &image.DisplayOrder, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func (r *PostImageRepository) DeleteByPostID(postID string) error {
	query := `DELETE FROM post_images WHERE food_post_id = $1`
	_, err := r.db.Exec(query, postID)
	return err
}

func (r *PostImageRepository) DeleteByID(imageID string) error {
	query := `DELETE FROM post_images WHERE id = $1`
	_, err := r.db.Exec(query, imageID)
	return err
}

func (r *PostImageRepository) GetImageURLsByPostID(postID string) ([]string, error) {
	var urls []string

	query := `
		SELECT image_url
		FROM post_images
		WHERE food_post_id = $1
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}
