package notifications

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	Id         uuid.UUID `json:"id" db:"id"`
	UserId     uuid.UUID `json:"user_id" db:"user_id"`
	Payload    string    `json:"payload" db:"payload"`
	LastSendAt time.Time `json:"-" db:"last_send_at"`
	Readed     bool      `json:"-" db:"readed"`
}
