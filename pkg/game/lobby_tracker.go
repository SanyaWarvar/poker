package game

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/google/uuid"
)

const (
	DefaultTimeout = time.Second * 15
	DefaultTTL     = time.Second * 30
	DefaultTTS     = time.Second * 5
)

type LobbyInfo struct {
	GameStarted  bool
	PlayersCount int
	MinPlayers   int
	LastActivity time.Time
	TTL          time.Duration
	TTS          time.Duration
}

type LobbyTracker struct {
	services IHoldemService
	lobbies  map[string]LobbyInfo
	mu       sync.RWMutex
	timeouts map[string]struct{}
}

var LobbyTrackerEventTypes = []string{"game_started", "next_move", "do", "game created", "game started", "stop_game"}

func NewLobbyTracker(s IHoldemService) *LobbyTracker {
	return &LobbyTracker{
		services: s,
		lobbies:  map[string]LobbyInfo{},
		mu:       sync.RWMutex{},
		timeouts: map[string]struct{}{},
	}
}

func (lt *LobbyTracker) Update(recipients []string, data holdem.ObserverMessage) {
	if !slices.Contains(LobbyTrackerEventTypes, data.EventType) {
		return
	}

	Id := data.LobbyId
	fmt.Println("lt", data.EventType, data.EventType == "stop_game")
	if data.EventType == "stop_game" {
		item := lt.lobbies[Id]
		item.GameStarted = false
		lt.lobbies[Id] = item
		go lt.GameMonitor(time.Second*5, Id)
	}

}

func (lt *LobbyTracker) GameMonitor(tts time.Duration, lobbyId string) {
	fmt.Println("monitor start")
	time.Sleep(tts)
	lobby, err := lt.services.GetLobbyById(uuid.MustParse(lobbyId))
	fmt.Println("monitor", lobby, err)
	if err != nil {
		return
	}
	fmt.Println("monitor", lt.lobbies[lobbyId].GameStarted, len(lobby.Players))
	if !lt.lobbies[lobbyId].GameStarted && len(lobby.Players) >= 2 {

		lt.services.StartGame(uuid.MustParse(lobbyId))
		item := lt.lobbies[lobbyId]
		item.GameStarted = true
		lt.lobbies[lobbyId] = item
	}

}

func (lt *LobbyTracker) AddPlayer(lId uuid.UUID) bool {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	l, ok := lt.lobbies[lId.String()]
	if !ok {
		return false
	}
	l.PlayersCount += 1
	l.LastActivity = time.Now()
	lt.lobbies[lId.String()] = l

	return true
}
