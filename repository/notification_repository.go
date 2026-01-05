package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/yourusername/food-sharing-backend/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notif *models.Notification) error {
	notif.ID = uuid.New().String()

	query := `
		INSERT INTO notifications (id, user_id, title, message, type, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	return r.db.QueryRow(
		query,
		notif.ID, notif.UserID, notif.Title, notif.Message, notif.Type, notif.ReferenceID,
	).Scan(&notif.CreatedAt)
}

func (r *NotificationRepository) FindByUserID(userID string) ([]models.Notification, error) {
	var notifs []models.Notification

	query := `
		SELECT id, user_id, title, message, type, reference_id, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Title, &notif.Message,
			&notif.Type, &notif.ReferenceID, &notif.IsRead, &notif.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifs = append(notifs, notif)
	}

	return notifs, nil
}

func (r *NotificationRepository) MarkAsRead(id string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *NotificationRepository) GetUnreadCount(userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}
