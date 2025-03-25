package game

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/google/uuid"
)

const pageSize = 50

var (
	ErrDuplicateLobbyId = errors.New("lobby id must be unique")
	ErrLobbyNotFound    = errors.New("lobby not found")
)

type IHoldemRepo interface {
	CreateLobby(cfg *holdem.TableConfig, lobbyId uuid.UUID) error
	GetLobbyList(page int) []holdem.TableConfig
	GetLobbyById(lobbyId uuid.UUID) (holdem.TableConfig, error)
	GetLobbyByPId(playerId uuid.UUID) (holdem.TableConfig, error)
	EnterInLobby(lobbyId uuid.UUID, player holdem.IPlayer) error
	OutFromLobby(lobbyId, playerId uuid.UUID) error
	DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error
	DeleteLobby(lobbyId uuid.UUID)
}

type HoldemRepo struct {
	db   map[string]holdem.IPokerTable
	list []string
	mu   *sync.RWMutex
}

func NewHoldemRepo() *HoldemRepo {
	return &HoldemRepo{
		db: make(map[string]holdem.IPokerTable),
		mu: &sync.RWMutex{},
	}
}

func (r *HoldemRepo) CreateLobby(cfg *holdem.TableConfig, lobbyId uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.db[lobbyId.String()]; ok {
		return ErrDuplicateLobbyId
	}
	r.db[lobbyId.String()] = holdem.NewPokerTable(cfg)
	r.list = append(r.list, lobbyId.String())
	fmt.Println(r.db, r.list)
	return nil
}

func (r *HoldemRepo) GetLobbyList(page int) []holdem.TableConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fmt.Println(r.db, r.list)
	start := page * 50
	end := (page + 1) * 50

	if start >= len(r.list) {
		return nil
	}
	if end > len(r.list) {
		end = len(r.list)
	}

	names := r.list[start:end]
	output := make([]holdem.TableConfig, 0, len(names))
	for _, v := range names {
		output = append(output, *r.db[v].GetConfig())
	}
	return output
}

func (r *HoldemRepo) GetLobbyById(lobbyId uuid.UUID) (holdem.TableConfig, error) {
	var output holdem.TableConfig
	r.mu.RLock()
	defer r.mu.RUnlock()
	table, ok := r.db[lobbyId.String()]
	if !ok {
		return output, ErrLobbyNotFound
	}
	output = *table.GetConfig()
	return output, nil
}

func (r *HoldemRepo) GetLobbyByPId(playerId uuid.UUID) (holdem.TableConfig, error) {
	var output holdem.TableConfig
	r.mu.RLock()
	defer r.mu.RUnlock()
	fmt.Println(r.db, r.list)
	for _, v := range r.db {
		fmt.Println(123123123123, v)
		if v.CheckPlayer(playerId.String()) {
			return *v.GetConfig(), nil
		}
	}
	return output, ErrLobbyNotFound
}

func (r *HoldemRepo) EnterInLobby(lobbyId uuid.UUID, player holdem.IPlayer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lobby, ok := r.db[lobbyId.String()]
	if !ok {
		return ErrLobbyNotFound
	}
	err := lobby.AddPlayer(player)
	return err
}

func (r *HoldemRepo) OutFromLobby(lobbyId, playerId uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lobby, ok := r.db[lobbyId.String()]
	if !ok {
		return ErrLobbyNotFound
	}
	err := lobby.RemovePlayer(playerId.String())
	return err
}

func (r *HoldemRepo) DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lobby, ok := r.db[lobbyId.String()]
	if !ok {
		return ErrLobbyNotFound
	}
	err := lobby.MakeMove(playerId.String(), action, amount)
	return err
}

func (r *HoldemRepo) DeleteLobby(lobbyId uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.db, lobbyId.String())
	ind := slices.Index(r.list, lobbyId.String())
	r.list = append(r.list[:ind], r.list[ind+1:]...)
}
