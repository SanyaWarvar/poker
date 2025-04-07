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
	service  IHoldemService
	Observer *WsObserver
	Lt       *LobbyTracker
}

func NewHoldemEngine(s IHoldemService, o *WsObserver, lt *LobbyTracker) *HoldemEngine {
	return &HoldemEngine{
		service:  s,
		Observer: o,
		Lt:       lt,
	}
}

func (e *HoldemEngine) StartEngine() {
}

func (e *HoldemEngine) NewLobby(lId, pId uuid.UUID, lInfo LobbyInfo) {
	e.Lt.lobbies[lId.String()] = lInfo
	go e.Lt.GameMonitor(time.Second*5, lId.String())
}

func (e *HoldemEngine) AddPlayer(lId uuid.UUID) bool {
	return e.Lt.AddPlayer(lId)
}

func (e *HoldemEngine) HandleMove(move PlayerMove) {
	e.service.DoAction(move.PlayerId, move.LobbyId, move.Action, move.Amount)
}
