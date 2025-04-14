package game

import (
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/gofiber/websocket/v2"
)

var WsObserverEventTypes = []string{}

type WsObserver struct {
	Conn map[string]*websocket.Conn
}

func NewWsObserver() *WsObserver {
	return &WsObserver{
		Conn: map[string]*websocket.Conn{},
	}
}

func (o *WsObserver) Update(recipients []string, data holdem.ObserverMessage) {

	o.Broadcast(recipients, data)

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
