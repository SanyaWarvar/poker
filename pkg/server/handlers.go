package server

import (
	"net/http"

	"strings"

	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) SignUp(c *fiber.Ctx) error {
	var input user.User

	err := c.BodyParser(&input)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if valid := input.IsValid(); !valid {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid username or password")
	}

	err = s.services.UserService.CreateUser(input)
	if err != nil {
		errorMessage := ""
		if strings.Contains(err.Error(), "email") {
			errorMessage = "This email already exist"
		}
		if strings.Contains(err.Error(), "username") {
			errorMessage = "This username already exist"
		}
		return ErrorResponse(c, http.StatusConflict, errorMessage)

	}
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"details": "ok",
	})
}
