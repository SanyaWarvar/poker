package game

import (
	"slices"

	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/gofiber/websocket/v2"
)

var WsObserverEventTypes = []string{"info", "game", "error"}

type WsObserver struct {
	Conn map[string]*websocket.Conn
}

func NewWsObserver() *WsObserver {
	return &WsObserver{
		Conn: map[string]*websocket.Conn{},
	}
}

func (o *WsObserver) Update(recipients []string, data holdem.ObserverMessage) {
	if slices.Contains(WsObserverEventTypes, data.EventType) {
		o.Broadcast(recipients, data)
	}
}

func (o *WsObserver) Broadcast(recipients []string, data holdem.ObserverMessage) {
	for _, recipient := range recipients {
		c, ok := o.Conn[recipient]
		if !ok {
			continue
		}
		c.WriteJSON(data)
	}
}
