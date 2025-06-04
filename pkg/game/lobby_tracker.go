package game

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

const (
	DefaultTimeout = time.Second * 15
	DefaultTTL     = time.Second * 30
	DefaultTTS     = time.Second * 10
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

	id := msg[1]

	if data.EventType == "stop_game" {
		lt.mu.Lock()
		item := lt.lobbies[id]
		item.GameStarted = false
		item.LastActivity = time.Now()
		lt.lobbies[id] = item
		lt.mu.Unlock()
	}
}

func (lt *LobbyTracker) GameMonitor(tts time.Duration) {
	for {
		currentTime := time.Now()
		lt.mu.Lock()
		for id, lobby := range lt.lobbies {
			if !lobby.GameStarted && lobby.LastActivity.Add(tts).Before(currentTime) && lobby.PlayersCount >= 2 {
				lt.mu.Unlock()
				lt.StartGame(id)
				lt.mu.Lock()
			}
		}
		lt.mu.Unlock()
	}
}

func (lt *LobbyTracker) StartGame(lobbyId string) {
	locked := lt.mu.TryLock()
	if !locked {
		fmt.Println("[ERROR] Failed to lock mutex in StartGame")
		return
	}
	item := lt.lobbies[lobbyId]
	item.LastActivity = time.Now()
	err := lt.services.StartGame(uuid.MustParse(lobbyId))
	item.GameStarted = true
	if err != nil {
		log.Warnf("StartGame: lt.services.StartGame: %s", err.Error())
		item.GameStarted = false
	}
	lt.lobbies[lobbyId] = item
	lt.mu.Unlock()
}

func (lt *LobbyTracker) AddPlayer(lId uuid.UUID) bool {
	lt.mu.Lock()
	l, ok := lt.lobbies[lId.String()]
	if !ok {
		lt.mu.Unlock()
		return false
	}
	l.PlayersCount += 1
	l.LastActivity = time.Now()
	lt.lobbies[lId.String()] = l
	lt.mu.Unlock()
	return true
}
