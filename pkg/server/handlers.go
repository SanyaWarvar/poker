package server

import (
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"time"

	"strings"

	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (s *Server) SignUp(c *fiber.Ctx) error {
	var input1 struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := c.BodyParser(&input1)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	input := user.User{Username: input1.Username, Email: input1.Email, Password: input1.Password}

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
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

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
		Secure:   true,
		Expires:  time.Now().Add(accessTokenTtl),
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(refreshTokenTtl),
		SameSite: "None",
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
		Secure:   true,
		Expires:  time.Now().Add(accessTokenTtl),
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(refreshTokenTtl),
		SameSite: "None",
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

func (s *Server) CheckAuthMiddleware(c *fiber.Ctx) error {
	accessTokenInput := c.Cookies("access_token")

	if accessTokenInput == "" {
		return ErrorResponse(c, http.StatusBadRequest, "Access token missing")
	}

	token, err := s.services.JwtService.ParseToken(accessTokenInput)

	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad access token")
	}
	c.Locals("userId", token.UserId)
	return c.Next()
}

func (s *Server) CheckAuthEndpoint(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"details": "Success",
	})
}

func ClearCookies(c *fiber.Ctx, key ...string) {
	for i := range key {
		c.Cookie(&fiber.Cookie{
			Name:    key[i],
			Expires: time.Now().Add(-time.Hour * 24),
			Value:   "",
		})
	}
}

func (s *Server) Logout(c *fiber.Ctx) error {
	ClearCookies(c, "access_token", "refresh_token")

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"details": "Success",
	})
}

func (s *Server) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"details": "username cant be empty"})
	}
	user, err := s.services.UserService.GetUserByUsername(username)
	if err != nil { // TODO возможно могут быть другие проблемы?
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"details": "user not found"})
	}
	user.GenerateUrl(c.Hostname())
	return c.Status(http.StatusOK).JSON(user)
}

func (s *Server) UpdateUserInfo(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
	}
	err := c.BodyParser(&input)
	if err != nil || !user.CheckUsername(input.Username) {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"details": "bad json"})
	}
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(map[string]string{"details": "bad user id"})
	}
	err = s.services.UserService.UpdateUsername(userId, input.Username)
	return c.Status(http.StatusNoContent).JSON(nil)
}

func (s *Server) UpdateProfilePic(c *fiber.Ctx) error {

	ProfilePic, err := c.FormFile("profile_pic")
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad form data")

	}
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	file, err := ProfilePic.Open()
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Unable to open file")
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Unable to open file")

	}
	suffix := filepath.Ext(ProfilePic.Filename)
	ValidFileSuffixForProfilePicture := []string{".gif", ".jpg", ".png"}
	if !slices.Contains(ValidFileSuffixForProfilePicture, suffix) {
		return ErrorResponse(c, http.StatusBadRequest, "Bad file format")
	}
	user, err := s.services.UserService.GetUserById(userId)
	err = s.services.UserService.UpdateProfilePic(userId, fileBytes, user.Username)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	return c.Status(http.StatusNoContent).JSON(nil)
}
