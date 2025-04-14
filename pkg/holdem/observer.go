package holdem

import (
	"fmt"
)

type IObserver interface {
	Update(recipients []string, data ObserverMessage)
}

type ObserverMessage struct {
	EventType string      `json:"event_type"`
	EventData interface{} `json:"event_data"`
	LobbyId   string      `json:"lobby_id"`
}

type Logger struct{}

func (l Logger) Update(recipients []string, data ObserverMessage) {
	fmt.Printf("[%s] %v\n", data.EventType, data.EventData)
}
