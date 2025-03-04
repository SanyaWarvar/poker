package server

import (
	"net/http"
	"time"

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

func (s *Server) SignIn(c *fiber.Ctx) error {
	var input user.User

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	user, err := s.services.UserService.GetUserByEP(input.Email, input.Password)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
	}

	isEmailConfirmed, _ := s.services.EmailSmtpService.CheckEmailConfirm(user.Email)

	if !isEmailConfirmed {
		return ErrorResponse(c, http.StatusForbidden, "Email not confirmed")
	}

	accessToken, refreshToken, _, err := s.services.JwtService.GeneratePairToken(user.Id)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	accessTokenTtl, refreshTokenTtl := s.services.JwtService.GetTokensTtl()

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(accessTokenTtl),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(refreshTokenTtl),
	})
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"details": "Sign in successful",
	})
}

func (s *Server) RefreshToken(c *fiber.Ctx) error {

	refreshTokenInput := c.Cookies("refresh_token")
	accessTokenInput := c.Cookies("access_token")
	if refreshTokenInput == "" {
		return ErrorResponse(c, http.StatusBadRequest, "Refresh token missing")
	}

	if accessTokenInput == "" {
		return ErrorResponse(c, http.StatusBadRequest, "Access token missing")
	}

	accessToken, err := s.services.JwtService.ParseToken(accessTokenInput)

	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad access token")
	}

	isTokenValid := s.services.JwtService.CheckRefreshTokenExp(accessToken.RefreshId)

	if !isTokenValid {
		return ErrorResponse(c, http.StatusUnauthorized, "bad refresh token")
	}

	newAccessToken, newRefreshToken, _, err := s.services.JwtService.GeneratePairToken(accessToken.UserId)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	accessTokenTtl, refreshTokenTtl := s.services.JwtService.GetTokensTtl()

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(accessTokenTtl),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(refreshTokenTtl),
	})

	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"details": "Token refreshed",
	})
}

// send code
// confirm code

func (s *Server) SendCode(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if input.Email == "" {
		return ErrorResponse(c, http.StatusBadRequest, "email missing")
	}

	if input.Password == "" {
		return ErrorResponse(c, http.StatusBadRequest, "password missing")
	}

	if err := s.services.EmailSmtpService.SendConfirmEmailMessage(input.Email); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"details": "Confirmation code sent",
	})
}

func (s *Server) ConfirmCode(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if input.Email == "" {
		return ErrorResponse(c, http.StatusBadRequest, "email missing")
	}

	if input.Code == "" {
		return ErrorResponse(c, http.StatusBadRequest, "code missing")
	}

	if err := s.services.EmailSmtpService.ConfirmEmail(input.Email, input.Code); err != nil {
		if err.Error() == "Bad code" {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid confirmation code")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to confirm code")
	}
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"message": "Email confirmed",
	})
}
