package game

import (
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/google/uuid"
)

type BalanceObserver struct {
	s user.IUserService
	h IHoldemService
}

func NewBalanceObserver(s user.IUserService, hs IHoldemService) *BalanceObserver {
	return &BalanceObserver{s: s, h: hs}
}

func (bo *BalanceObserver) Update(recipients []string, data holdem.ObserverMessage) {
	if data.EventType == "players_stats" {
		players, ok := data.EventData.([]holdem.IPlayer)
		if !ok {
			return
		}
		ids := make([]uuid.UUID, 0, len(players))
		balance := make([]int, 0, len(players))
		for _, p := range players {
			ids = append(ids, uuid.MustParse(p.GetId()))
			balance = append(balance, p.GetBalance())
		}
		bo.s.UpdateManyUserBalance(ids, balance)
	}

	/*if data.EventType == "stop_game" {
		l, err := bo.h.GetLobbyById(uuid.MustParse(data.LobbyId))
		if err != nil {
			return
		}
		for _, u := range l.Players {
			bo.s.IncGameCount(u.Id)
			bo.s.UpdateMaxBalance(u.Id)
		}
	}*/
}
