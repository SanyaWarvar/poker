package handlers

import (
	"encoding/json"

	_ "github.com/SanyaWarvar/poker/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

// ErrorResponse
// @Schema
type ErrorResponseStruct struct {
	Message string `json:"message"`
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	logrus.Error(message)
	return c.Status(statusCode).JSON(ErrorResponseStruct{Message: message})
}

func WsErrorResponse(c *websocket.Conn, messageType int, message string) error {
	logrus.Error(message)
	data, err := json.Marshal(ErrorResponseStruct{Message: message})
	if err != nil {
		return err
	}
	return c.WriteMessage(messageType, data)
}

func PingHandler(c *websocket.Conn) {
	c.WriteMessage(websocket.PongMessage, []byte("ping"))
}
