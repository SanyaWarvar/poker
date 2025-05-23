package notifications

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationsPostgres struct {
	db *sqlx.DB
}

type INotificationRepository interface {
	MarkReaded(notificationId, userId uuid.UUID) error
	CreateNotification(item Notification) error
	GetNotReadedNotifiesByUserId(userId uuid.UUID) ([]Notification, error)
}

func NewNotificationsPostgres(db *sqlx.DB) *NotificationsPostgres {
	return &NotificationsPostgres{db: db}
}

func (r *NotificationsPostgres) GetNotReadedNotifiesByUserId(userId uuid.UUID) ([]Notification, error) {
	query := `
		UPDATE notifications SET last_send_at = $1
		WHERE user_id = $2 and readed = 'f' and last_send_at + '30 second'::interval < $1
		RETURNING *
	`
	var output []Notification
	err := r.db.Select(&output, query, time.Now(), userId)
	return output, err
}

func (r *NotificationsPostgres) CreateNotification(item Notification) error {
	query := `
		INSERT INTO notifications(id, user_id, payload, last_send_at, readed) VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, item.Id, item.UserId, item.Payload, time.Now(), false)
	return err
}

func (r *NotificationsPostgres) MarkReaded(notificationId, userId uuid.UUID) error {
	query := `
		UPDATE notifications SET readed = 't' WHERE id = $1 and user_id = $2;
	`
	_, err := r.db.Exec(query, notificationId, userId)
	return err
}
