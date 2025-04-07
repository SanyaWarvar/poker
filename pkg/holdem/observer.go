package holdem

import (
	"fmt"
)

type IObserver interface {
	Update(recipients []string, data ObserverMessage)
}

type ObserverMessage struct {
	EventType string
	EventData string
}

type Logger struct{}

func (l Logger) Update(recipients []string, data ObserverMessage) {
	fmt.Printf("[%s] %v\n", data.EventType, data.EventData)
}
