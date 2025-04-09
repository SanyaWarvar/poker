package game

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

var LobbyTrackerEventTypes = []string{"info"}

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
	if len(msg) > 4 && strings.Join(msg[0:4], " ") == "Next move expect from" {
		Id := msg[4]
		lt.timeouts[Id] = struct{}{}
		go lt.TurnTimeout(Id, data.LobbyId)
	}
	Id := msg[1]
	if len(msg) > 4 && strings.Join([]string{msg[0], msg[2]}, " ") == "player do" {
		log.Printf("timeout delete for %s", Id)
		delete(lt.timeouts, Id)
	}

	if (len(msg) == 5 && strings.Join(slices.Delete(msg, 1, 2), " ") == "game has been stopped") ||
		(len(msg) == 3 && strings.Join(slices.Delete(msg, 1, 2), " ") == "game started") ||
		(len(msg) == 3 && strings.Join(slices.Delete(msg, 1, 2), " ") == "game created") {
		log.Printf("timeout add for %s", Id)
		go lt.GameMonitor(time.Second*1, Id)
	}

}

func (lt *LobbyTracker) TurnTimeout(playerId, lobbyId string) {
	time.Sleep(DefaultTimeout)
	lt.mu.Lock()
	defer lt.mu.Unlock()
	_, ok := lt.timeouts[playerId]
	if ok {
		//lt.services.DoAction(uuid.MustParse(playerId), uuid.MustParse(lobbyId), "fold", 0)
	}
}

func (lt *LobbyTracker) GameMonitor(interval time.Duration, lobbyId string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	lt.mu.Lock()
	info, ok := lt.lobbies[lobbyId]
	fmt.Println(ok, lt.lobbies)
	if !ok {
		return
	}
	info.GameStarted = false
	info.LastActivity = time.Now()
	lt.lobbies[lobbyId] = info
	lt.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			logrus.Printf("lobby %s monitoring stopped: %v", lobbyId, ctx.Err())
			return
		case <-ticker.C:
			lt.mu.Lock()
			info, exists := lt.lobbies[lobbyId]
			if !exists {
				lt.mu.Unlock()
				return
			}

			if !info.GameStarted && info.PlayersCount >= info.MinPlayers {
				go func(id string, tts time.Duration) {
					time.Sleep(tts)
					lt.mu.Lock()
					defer lt.mu.Unlock()
					if info, exists := lt.lobbies[id]; exists &&
						!info.GameStarted &&
						info.PlayersCount >= info.MinPlayers {
						info.GameStarted = true
						lt.lobbies[id] = info
						lt.services.StartGame(uuid.MustParse(id))
						cancel()
					}
				}(lobbyId, info.TTS)

			} else if !info.GameStarted && info.PlayersCount == 0 {
				go func(id string, ttl time.Duration) {
					time.Sleep(ttl)
					lt.mu.Lock()
					defer lt.mu.Unlock()

					info, exists := lt.lobbies[id]
					if exists && !info.GameStarted && info.PlayersCount == 0 {
						logrus.Printf("lobby %s deleted due to inactivity", id)
						lt.services.DeleteLobby(uuid.MustParse(id))
						cancel()
					}
				}(lobbyId, info.TTL)
			}
			lt.mu.Unlock()
		}
	}
}

func (lt *LobbyTracker) StartBackgroundTasks() {
	//go lt.cleanupEmptyLobbiesMonitor(time.Second)
	//go lt.startGameMonitor(time.Second)
}

func (lt *LobbyTracker) AddPlayer(lId uuid.UUID) bool {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	l, ok := lt.lobbies[lId.String()]
	fmt.Println(lt.lobbies)
	if !ok {
		return false
	}
	l.PlayersCount += 1
	l.LastActivity = time.Now()
	lt.lobbies[lId.String()] = l
	fmt.Println(lt.lobbies[lId.String()])
	return true
}
