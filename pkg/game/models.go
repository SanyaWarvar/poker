package game

import (
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/SanyaWarvar/poker/pkg/user"
)

type LobbyOutput struct {
	Info    holdem.TableConfig `json:"info"`
	Players []user.User        `json:"players"`
}
