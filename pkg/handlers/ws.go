package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

func (h *Handler) EnterInLobby(c *websocket.Conn) {
	userId, err := h.WSGetUserId(c)
	fmt.Println(err)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, "bad token")
		return
	}
	lobbyID, err := uuid.Parse(c.Query("lobby_id"))
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, "no or invalid lobby id")
		return
	}
	_, err = h.services.HoldemService.GetLobbyById(lobbyID)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	user, err := h.services.UserService.GetUserById(userId)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	err = h.services.HoldemService.EnterInLobby(lobbyID, userId, user.Balance)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	c.WriteJSON(map[string]string{"details": "success"})

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("Ping error:", err)
					return
				}
				log.Println("Sent ping")
			case <-done:
				return
			}
		}
	}()

	c.SetPongHandler(func(string) error {
		log.Println("Received pong")
		return nil
	})

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			close(done)
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Client disconnected: %v", err)
			}
			return
		}

		log.Printf("Received: %s", msg)

		if err := c.WriteMessage(messageType, msg); err != nil {
			close(done)
			log.Println("Write error:", err)
			return
		}
	}
}
