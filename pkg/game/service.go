package game

import (
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/google/uuid"
)

type IHoldemService interface {
	CreateLobby(cfg *holdem.TableConfig, playerId uuid.UUID) (uuid.UUID, error)
	GetLobbyList(page int) ([]LobbyOutput, error)
	GetLobbyById(lobbyId uuid.UUID) (LobbyOutput, error)
	GetLobbyByPId(playerId uuid.UUID) (LobbyOutput, error)
	EnterInLobby(lobbyId, playerId uuid.UUID, balance int) error
	OutFromLobby(lobbyId, playerId uuid.UUID) error
	DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error
	AddObserver(lobbyId uuid.UUID, observer holdem.IObserver) error
	StartGame(lobbyId uuid.UUID) error
	DeleteLobby(lobbyId uuid.UUID)
}

type HoldemService struct {
	holdemRepo IHoldemRepo
	userRepo   user.IUserRepo
}

func NewHoldemService(holdemRepo IHoldemRepo, userRepo user.IUserRepo) *HoldemService {
	return &HoldemService{
		holdemRepo: holdemRepo,
		userRepo:   userRepo,
	}
}

func (s *HoldemService) CreateLobby(cfg *holdem.TableConfig, playerId uuid.UUID) (uuid.UUID, error) {

	var lobbyId uuid.UUID
	for {
		lobbyId = uuid.New()
		cfg.TableId = lobbyId
		err := s.holdemRepo.CreateLobby(cfg, lobbyId)
		if err == ErrDuplicateLobbyId {
			continue
		}
		return lobbyId, nil
	}
}

func (s *HoldemService) GetLobbyList(page int) ([]LobbyOutput, error) {
	info := s.holdemRepo.GetLobbyList(page)
	output := make([]LobbyOutput, 0, len(info))
	for ind, _ := range info {
		playersId, err := s.holdemRepo.PlayersIdFromLobbyById(info[ind].TableId)
		if err != nil {
			return []LobbyOutput{}, err
		}
		players, err := s.userRepo.GetPlayersByIdLIst(playersId)
		if err != nil {
			return []LobbyOutput{}, err
		}
		output = append(output, LobbyOutput{
			Info:    info[ind],
			Players: players,
		})

	}
	return output, nil
}

func (s *HoldemService) GetLobbyById(lobbyId uuid.UUID) (LobbyOutput, error) {
	var output LobbyOutput
	info, err := s.holdemRepo.GetLobbyById(lobbyId)
	if err != nil {
		return output, err
	}
	pId, err := s.holdemRepo.PlayersIdFromLobbyById(lobbyId)
	if err != nil {
		return output, err
	}
	players, err := s.userRepo.GetPlayersByIdLIst(pId)
	if err != nil {
		return output, err
	}
	return LobbyOutput{Info: info, Players: players}, nil
}

func (s *HoldemService) GetLobbyByPId(playerId uuid.UUID) (LobbyOutput, error) {
	var output LobbyOutput

	info, err := s.holdemRepo.GetLobbyByPId(playerId)
	if err != nil {
		return output, err
	}
	pId, err := s.holdemRepo.PlayersIdFromLobbyById(info.TableId)
	if err != nil {
		return output, err
	}
	players, err := s.userRepo.GetPlayersByIdLIst(pId)
	if err != nil {
		return output, err
	}
	return LobbyOutput{Info: info, Players: players}, nil
}

// TODO change this
func (s *HoldemService) EnterInLobby(lobbyId, playerId uuid.UUID, balance int) error {
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
	if lobby.Info.BankAmount != 0 {
		p.Balance = lobby.Info.BankAmount
	}
	return s.holdemRepo.EnterInLobby(lobbyId, p)
}

func (s *HoldemService) OutFromLobby(lobbyId, playerId uuid.UUID) error {
	return s.holdemRepo.OutFromLobby(lobbyId, playerId)
}

func (s *HoldemService) AddObserver(lobbyId uuid.UUID, observer holdem.IObserver) error {
	return s.holdemRepo.AddObserver(lobbyId, observer)
}

func (s *HoldemService) DoAction(playerId, lobbyId uuid.UUID, action string, amount int) error {
	return s.holdemRepo.DoAction(playerId, lobbyId, action, amount)
}

func (s *HoldemService) StartGame(lobbyId uuid.UUID) error {
	return s.holdemRepo.StartGame(lobbyId)
}
func (s *HoldemService) DeleteLobby(lobbyId uuid.UUID) {
	s.holdemRepo.DeleteLobby(lobbyId)
}
