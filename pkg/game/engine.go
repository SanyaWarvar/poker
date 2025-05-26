package game

import (
	"time"

	"github.com/google/uuid"
)

type PlayerMove struct {
	PlayerId uuid.UUID
	LobbyId  uuid.UUID
	Action   string `json:"action" binding:"reqired"`
	Amount   int    `json:"amount" binding:"reqired"`
}

type HoldemEngine struct {
	service    IHoldemService
	WsObserver *WsObserver
	BObserver  *BalanceObserver
	Lt         *LobbyTracker
}

func NewHoldemEngine(s IHoldemService, o *WsObserver, b *BalanceObserver, lt *LobbyTracker) *HoldemEngine {
	return &HoldemEngine{
		service:    s,
		WsObserver: o,
		BObserver:  b,
		Lt:         lt,
	}
}

func (e *HoldemEngine) NewLobby(lId, pId uuid.UUID, lInfo LobbyInfo) {
	e.Lt.lobbies[lId.String()] = lInfo
	go e.Lt.GameMonitor(time.Second*5, lId.String())
}

func (e *HoldemEngine) AddPlayer(lId, pId uuid.UUID) bool {
	go e.Lt.GameMonitor(DefaultTTS, lId.String())
	return e.Lt.AddPlayer(lId)

}

func (e *HoldemEngine) HandleMove(move PlayerMove) {
	e.service.DoAction(move.PlayerId, move.LobbyId, move.Action, move.Amount)
}
