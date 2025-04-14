package game

import (
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/google/uuid"
)

type BalanceObserver struct {
	s user.IUserService
}

func NewBalanceObserver(s user.IUserService) *BalanceObserver {
	return &BalanceObserver{s: s}
}

func (bo *BalanceObserver) Update(recipients []string, data holdem.ObserverMessage) {
	if data.EventType != "players_stats" {
		return
	}
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
