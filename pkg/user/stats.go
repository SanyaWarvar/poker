package user

import "github.com/google/uuid"

type PlayerStats struct {
	UserId     uuid.UUID `json:"-" db:"user_id"`
	MaxBalance int       `json:"max_balance" db:"max_balance"`
	GameCount  int       `json:"game_count" db:"games_played"`
}
