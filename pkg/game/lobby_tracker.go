package game

import (
	"log"
	"slices"
	"strings"
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
	s, ok := data.EventData.(string)
	if !ok {
		return
	}
	msg := strings.Split(s, " ")

	Id := msg[1]
	if data.EventType == "do" {
		log.Printf("timeout delete for %s", Id)
		delete(lt.timeouts, Id)
	}

	if data.EventType == "stop_game" || data.EventType == "game started" || data.EventType == "game created" {
		go lt.GameMonitor(time.Second*1, Id)
	}

}

func (lt *LobbyTracker) GameMonitor(tts time.Duration, lobbyId string) {
	time.Sleep(tts)
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lobby, err := lt.services.GetLobbyById(uuid.MustParse(lobbyId))
	if err != nil {
		return
	}
	if !lt.lobbies[lobbyId].GameStarted && len(lobby.Players) >= 2 {
		lt.services.StartGame(uuid.MustParse(lobbyId))
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
