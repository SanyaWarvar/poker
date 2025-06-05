package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"

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
		fmt.Println("Received pong")
		return nil
	})
	go h.handleDisconnect(c, userId, lobbyID, done)
	for {
		var pMove game.PlayerMove
		_, msg, err := c.ReadMessage()
		if err != nil {
			close(done)
			if websocket.IsUnexpectedCloseError(err) {
				fmt.Printf("Client disconnected: %v\n", err)
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

func (h *Handler) handleDisconnect(c *websocket.Conn, userId uuid.UUID, lobbyID uuid.UUID, done chan struct{}) {
	<-done

	delete(h.engine.WsObserver.Conn, userId.String())

	err := h.engine.OutFromLobby(lobbyID, userId)
	if err != nil {
		log.Warnf("handleDisconnect: h.services.HoldemService.OutFromLobby: %s", err.Error())
	}
}
