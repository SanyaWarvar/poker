package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/SanyaWarvar/poker/docs"
	"github.com/SanyaWarvar/poker/pkg/game"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

func (h *Handler) EnterInLobby(c *websocket.Conn) {
	_, msg, err := c.ReadMessage()
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	token, err := h.services.JwtService.ParseToken(string(msg), true)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	userId := token.UserId
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
	h.engine.WsObserver.Conn[userId.String()] = c
	ok := h.engine.AddPlayer(lobbyID, userId)
	if !ok {
		WsErrorResponse(c, websocket.CloseMessage, "cant enter")
		return
	}
	lInfo, err := h.services.HoldemService.GetLobbyById(lobbyID)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	for ind, v := range lInfo.Players {
		v.GenerateUrl()
		lInfo.Players[ind] = v
	}
	c.WriteJSON(lInfo)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
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
		var pMove game.PlayerMove
		_, msg, err := c.ReadMessage()
		if err != nil {
			close(done)
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Client disconnected: %v", err)
			}
			return
		}
		err = json.Unmarshal(msg, &pMove)
		if err != nil {
			WsErrorResponse(c, websocket.TextMessage, err.Error())
		}
		pMove.PlayerId = userId
		pMove.LobbyId = lobbyID
		h.engine.HandleMove(pMove)
	}
}

func (h *Handler) NotificationsConnect(c *websocket.Conn) {
	_, msg, err := c.ReadMessage()
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	token, err := h.services.JwtService.ParseToken(string(msg), true)
	if err != nil {
		WsErrorResponse(c, websocket.CloseMessage, err.Error())
		return
	}
	userId := token.UserId
	go func(userId uuid.UUID) {
		for {
			time.Sleep(time.Second * 1)
			output, err := h.services.NotificationService.GetNotReadedNotifiesByUserId(userId)

			if err != nil || len(output) == 0 {
				continue
			}
			if len(output) == 0 {
				continue
			}
			err = c.WriteJSON(output)
			if err != nil {
				WsErrorResponse(c, websocket.CloseMessage, err.Error())
				return
			}
		}
	}(userId)

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
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

		var id struct {
			Id uuid.UUID `json:"id"`
		}
		_, msg, err := c.ReadMessage()
		if err != nil {
			close(done)
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Client disconnected: %v", err)
			}
			return
		}
		err = json.Unmarshal(msg, &id)
		if err != nil {
			WsErrorResponse(c, websocket.TextMessage, err.Error())
		}
		fmt.Println(userId, "read", id.Id)
		err = h.services.NotificationService.MarkReaded(id.Id, userId)
		if err != nil {
			fmt.Println(userId, "err", err.Error())
		}

	}
}
