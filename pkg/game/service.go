package game

import (
	"fmt"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/google/uuid"
)

type IHoldemService interface {
	CreateLobby(cfg *holdem.TableConfig, playerId uuid.UUID) (uuid.UUID, error)
	GetLobbyList(page int) []holdem.TableConfig
	GetLobbyById(lobbyId uuid.UUID) (holdem.TableConfig, error)
	GetLobbyByPId(playerId uuid.UUID) (holdem.TableConfig, error)
	EnterInLobby(lobbyId, playerId uuid.UUID, balance int) error
	OutFromLobby(lobbyId, playerId uuid.UUID) error
	DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error
}

type HoldemService struct {
	repo IHoldemRepo
}

func NewHoldemService(repo IHoldemRepo) *HoldemService {
	return &HoldemService{
		repo: repo,
	}
}

func (s *HoldemService) CreateLobby(cfg *holdem.TableConfig, playerId uuid.UUID) (uuid.UUID, error) {
	var lobbyId uuid.UUID
	for {
		lobbyId = uuid.New()
		cfg.TableId = lobbyId
		err := s.repo.CreateLobby(cfg, lobbyId)
		if err == ErrDuplicateLobbyId {
			continue
		}
		fmt.Println(lobbyId)
		return lobbyId, nil
	}
}

func (s *HoldemService) GetLobbyList(page int) []holdem.TableConfig {
	return s.repo.GetLobbyList(page)
}

func (s *HoldemService) GetLobbyById(lobbyId uuid.UUID) (holdem.TableConfig, error) {
	return s.repo.GetLobbyById(lobbyId)
}

func (s *HoldemService) GetLobbyByPId(playerId uuid.UUID) (holdem.TableConfig, error) {
	return s.repo.GetLobbyByPId(playerId)
}

func (s *HoldemService) EnterInLobby(lobbyId, playerId uuid.UUID, balance int) error { //TODO change this
	lobby, err := s.GetLobbyById(lobbyId)
	if err != nil {
		return err
	}
	p := &holdem.Player{
		Id:      playerId,
		Balance: balance,
		Status:  false,
		LastBet: 0,
		Hand:    holdem.Hand{Cards: [2]holdem.Card{}},
		IsFold:  false,
	}
	if lobby.BankAmount != 0 {
		p.Balance = lobby.BankAmount
	}
	return s.repo.EnterInLobby(lobbyId, p)
}

func (s *HoldemService) OutFromLobby(lobbyId, playerId uuid.UUID) error {
	return s.repo.OutFromLobby(lobbyId, playerId)
}

func (s *HoldemService) DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error {
	return s.repo.DoAction(playerId, lobbyId, action, amount)
}
