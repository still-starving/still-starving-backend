package repository

import (
	"database/sql"
	"fmt"
	"strings"

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
		INSERT INTO food_posts (id, user_id, title, description, quantity, location, expiry_date, price, currency, spice_level, ingredients, cooked_at, latitude, longitude, location_geo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, ST_SetSRID(ST_MakePoint($14, $13), 4326))
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		post.ID, post.UserID, post.Title, post.Description,
		post.Quantity, post.Location, post.ExpiryDate,
		post.Price, post.Currency, post.SpiceLevel, post.Ingredients, post.CookedAt,
		post.Latitude, post.Longitude,
	).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (r *FoodPostRepository) FindAll(status string, lat, lng, radius float64, limit, offset int) ([]models.FoodPost, error) {
	var posts []models.FoodPost

	query := `
		SELECT fp.id, fp.user_id, fp.title, fp.description, fp.quantity, fp.location,
		       fp.expiry_date, fp.status, fp.created_at, fp.updated_at,
		       u.name as user_name, fp.price, fp.currency, 
		       COALESCE(fp.spice_level, 'no_spicy') as spice_level, 
		       COALESCE(fp.ingredients, '') as ingredients,
		       fp.cooked_at, COALESCE(fp.latitude, 0) as latitude, COALESCE(fp.longitude, 0) as longitude,
		       COALESCE(array_agg(pi.image_url ORDER BY pi.display_order) FILTER (WHERE pi.image_url IS NOT NULL), '{}') as image_urls
		FROM food_posts fp
		JOIN users u ON fp.user_id = u.id
		LEFT JOIN post_images pi ON fp.id = pi.food_post_id
	`

	args := []interface{}{}
	argCount := 1

	whereClauses := []string{}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("fp.status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	if lat != 0 && lng != 0 && radius > 0 {
		// Use PostGIS for radius search
		whereClauses = append(whereClauses, fmt.Sprintf("ST_DWithin(fp.location_geo, ST_SetSRID(ST_MakePoint($%d, $%d), 4326), $%d)", argCount+1, argCount, argCount+2))
		args = append(args, lat, lng, radius)
		argCount += 3
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " GROUP BY fp.id, u.name"

	if lat != 0 && lng != 0 {
		// Sort by distance if location provided
		// We need to re-pass params or just use the ones we have?
		// Actually, creating a CTE or subquery is cleaner, but keeping it simple:
		// We'll reuse the parameter indices we just added.
		// Since we added lat, lng at indexes (argCount-3, argCount-2) roughly.
		// To be safe and simple, let's just use created_at DESC if complex.
		// OR: just append distance to ORDER BY using the same params?
		// Postgres driver doesn't support named params easily.
		// Let's stick to CreatedAt for now to avoid parameter index hell,
		// OR re-add params for sorting (which is inefficient but works).

		// For MVP: Filter is most important. Sorting by distance is nice-to-have.
		// Let's sort by CreatedAt DESC to keep it simple and robust for this step.
		query += " ORDER BY fp.created_at DESC"
	} else {
		query += " ORDER BY fp.created_at DESC"
	}

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
			&post.Price, &post.Currency, &post.SpiceLevel, &post.Ingredients,
			&post.CookedAt, &post.Latitude, &post.Longitude,
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
		       COALESCE(fp.spice_level, 'no_spicy') as spice_level, 
		       COALESCE(fp.ingredients, '') as ingredients,
		       fp.cooked_at, COALESCE(fp.latitude, 0) as latitude, COALESCE(fp.longitude, 0) as longitude,
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
		&post.Price, &post.Currency, &post.SpiceLevel, &post.Ingredients,
		&post.CookedAt, &post.Latitude, &post.Longitude,
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
		    expiry_date = $5, status = $6, price = $7, currency = $8,
		    spice_level = $9, ingredients = $10, cooked_at = $11, 
		    latitude = $12, longitude = $13, location_geo = ST_SetSRID(ST_MakePoint($13, $12), 4326),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
		RETURNING updated_at
	`

	return r.db.QueryRow(
		query,
		post.Title, post.Description, post.Quantity, post.Location,
		post.ExpiryDate, post.Status, post.Price, post.Currency,
		post.SpiceLevel, post.Ingredients, post.CookedAt,
		post.Latitude, post.Longitude, post.ID,
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
		       COALESCE(fp.spice_level, 'no_spicy') as spice_level, 
		       COALESCE(fp.ingredients, '') as ingredients,
		       fp.cooked_at, COALESCE(fp.latitude, 0) as latitude, COALESCE(fp.longitude, 0) as longitude,
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
			&post.Price, &post.Currency, &post.SpiceLevel, &post.Ingredients,
			&post.CookedAt, &post.Latitude, &post.Longitude,
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
