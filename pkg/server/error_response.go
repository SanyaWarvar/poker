package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ErrorResponse
// @Schema
type ErrorResponseStruct struct {
	Message string `json:"message"`
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	logrus.Error(message)
	c.Status(statusCode)
	return c.JSON(ErrorResponseStruct{Message: message})
}
