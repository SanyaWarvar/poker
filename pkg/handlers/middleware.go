package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

func (h *Handler) CheckAuthMiddleware(c *fiber.Ctx) error {
	accessToken := strings.Split(c.Get("Authorization"), " ")[1]
	if accessToken == "" {
		return ErrorResponse(c, http.StatusBadRequest, "Access token missing")
	}

	token, err := h.services.JwtService.ParseToken(accessToken, true)

	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad access token")
	}
	c.Locals("userId", token.UserId)
	return c.Next()
}

func (h *Handler) WSGetUserId(conn *websocket.Conn) (uuid.UUID, error) {
	var output uuid.UUID
	authHeader := conn.Headers("Authorization")
	if len(authHeader) == 0 {
		return output, errors.New("Authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return output, errors.New("Invalid auth format")
	}
	accessToken := parts[1]

	token, err := h.services.JwtService.ParseToken(accessToken, true)
	if err != nil {
		return output, errors.New("Invalid token")
	}
	output = token.UserId
	return output, nil
}
