package server

import (
	"fmt"
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

// UserInput
// @Schema
type UserInput struct {
	Username string `json:"username" example:"john_doe" binding:"reqired"`
	Email    string `json:"email" example:"john@example.com" binding:"reqired"`
	Password string `json:"password" example:"password" binding:"reqired"`
}

// SignUp
// @Summary Регистрирация
// @Description Регистрирует нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param body body UserInput true "Данные пользователя"
// @Success 201 {object} map[string]string "Успешный ответ"
// @Failure 400 {object} ErrorResponseStruct "Invalid username or password"
// @Failure 404 {object} ErrorResponseStruct "This email already exist"
// @Failure 404 {object} ErrorResponseStruct "This username already exist"
// @Router /auth/sign_up [post]
func (s *Server) SignUp(c *fiber.Ctx) error {
	var input1 UserInput

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
	return c.JSON(map[string]string{
		"details": "ok",
	})
}

// EmailAndPasswordInput
// @Schema
type EmailAndPasswordInput struct {
	Email    string `json:"email" example:"john@example.com" binding:"reqired"`
	Password string `json:"password" example:"password" binding:"reqired"`
}

// SignIn
// @Summary Вход
// @Description Вход в аккаунт с подтвержденной почтой
// @Tags auth
// @Accept json
// @Produce json
// @Param body body EmailAndPasswordInput true "Данные пользователя"
// @Success 201 {object} user.user "Успешный ответ"
// @Header 201 {string} SetCookie1 "access_token secure=true http_only=true"
// @Header 201 {string} SetCookie2 "refresh_token secure=true http_only=true"
// @Failure 400 {object} ErrorResponseStruct "Invalid json"
// @Failure 401 {object} ErrorResponseStruct "Invalid email or password"
// @Failure 403 {object} ErrorResponseStruct "Email not confirmed""
// @Failure 500 {object} ErrorResponseStruct "Failed to generate tokens"
// @Router /auth/sign_in [post]
func (s *Server) SignIn(c *fiber.Ctx) error {
	var input EmailAndPasswordInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	user, err := s.services.UserService.GetUserByEP(input.Email, input.Password)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
	}
	user.GenerateUrl(c.Hostname())
	if !user.IsEmailConfirmed {
		return ErrorResponse(c, http.StatusForbidden, "Email not confirmed")
	}

	accessToken, refreshToken, _, err := s.services.JwtService.GeneratePairToken(user.Id)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	accessTokenTtl, refreshTokenTtl := s.services.JwtService.GetTokensTtl()

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Domain:   "http://localhost:5173",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(accessTokenTtl),
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Domain:   "http://localhost:5173",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(refreshTokenTtl),
		SameSite: "None",
	})
	c.Status(http.StatusCreated)
	return c.JSON(user)
}

// RefreshToken
// @Summary Обновление токенов
// @Description Обновляет хедеры с авторизационными токенами
// @Tags auth
// @Produce json
// @Success 201 {object} map[string]string "Успешный ответ"
// @Header 201 {string} SetCookie "access_token secure=true http_only=true"
// @Header 201 {string} SetCookie "refresh_token secure=true http_only=true"
// @Failure 400 {object} ErrorResponseStruct "Refresh token missing"
// @Failure 400 {object} ErrorResponseStruct "Access token missing"
// @Failure 400 {object} ErrorResponseStruct "Bad access token"
// @Failure 401 {object} ErrorResponseStruct "Bad refresh token"
// @Failure 500 {object} ErrorResponseStruct "Failed to generate tokens"
// @Router /auth/refresh_token [post]
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
		Domain:   "http://localhost:5173",
		Value:    newAccessToken,
		HTTPOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(accessTokenTtl),
		SameSite: "None",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Domain:   "http://localhost:5173",
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

// SendCode
// @Summary Отрпавить код
// @Description Отправляет код подтверждения почты
// @Tags auth
// @Accept json
// @Produce json
// @Param body body EmailAndPasswordInput true "Данные пользователя"
// @Success 201 {object} map[string]string "Успешный ответ"
// @Failure 400 {object} ErrorResponseStruct "invalid json"
// @Failure 400 {object} ErrorResponseStruct "email already confirmed"
// @Router /auth/send_code [post]
func (s *Server) SendCode(c *fiber.Ctx) error {
	var input EmailAndPasswordInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if err := s.services.EmailSmtpService.SendConfirmEmailMessage(input.Email); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"details": "Confirmation code sent",
	})
}

type ConfirmCodeInput struct {
	Email string `json:"email" binding:"reqired" example:"john@example.com"`
	Code  string `json:"code" binding:"reqired" example:"123456"`
}

// ConfirmCode
// @Summary Подтвердить почту
// @Description Подтвердить почту
// @Tags auth
// @Accept json
// @Produce json
// @Param body body ConfirmCodeInput true "Данные пользователя"
// @Success 201 {object} map[string]string "Успешный ответ"
// @Failure 400 {object} ErrorResponseStruct "invalid json"
// @Failure 400 {object} ErrorResponseStruct "email already confirmed"
// @Failure 400 {object} ErrorResponseStruct "Invalid confirmation code"
// @Failure 500 {object} ErrorResponseStruct "Failed to confirm code"
// @Router /auth/confirm_email [post]
func (s *Server) ConfirmCode(c *fiber.Ctx) error {
	var input ConfirmCodeInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
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

// CheckAuth
// @Summary Проверка валидности токенов
// @Description Проверка содержится ли в куках валидный токен доступа
// @Security ApiAuth
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Успешный ответ"
// @Failure 401 {object} ErrorResponseStruct "access token missing"
// @Failure 401 {object} ErrorResponseStruct "bad access token"
// @Router /auth/check_auth [post]
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

// Logout
// @Summary Выйти из аккаунта
// @Description Очищает все куки (токены)
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Успешный ответ"
// @Router /auth/logout [post]
func (s *Server) Logout(c *fiber.Ctx) error {
	ClearCookies(c, "access_token", "refresh_token")

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"details": "Success",
	})
}

// GetUser
// @Summary Получить пользователя по имени
// @Description Возвращает данные пользователя по его имени.
// @Security ApiAuth
// @Tags user
// @Produce json
// @Param username path string true "Имя пользователя"
// @Success 200 {object} user.User "Успешный ответ"
// @Failure 400 {object} map[string]string "username cant be empty"
// @Failure 404 {object} map[string]string "user not found"
// @Router /user/{username} [get]
func (s *Server) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")
	fmt.Println(username)
	if username == "" {
		return ErrorResponse(c, http.StatusBadRequest, "username cant be empty")
	}
	user, err := s.services.UserService.GetUserByUsername(username)
	if err != nil { // TODO возможно могут быть другие проблемы?
		return ErrorResponse(c, http.StatusNotFound, "user not found")
	}
	user.GenerateUrl(c.Hostname())
	return c.Status(http.StatusOK).JSON(user)
}

type UsernameInput struct {
	Username string `json:"username" binding:"reqired" example:"john doe"`
}

// UpdateUserInfo
// @Summary Обновить пользовательские данные
// @Description Обновляет username пользователя
// @Security ApiAuth
// @Tags user
// @Accept json
// @Produce json
// @Param body body UsernameInput true "Данные пользователя"
// @Success 204 {string} string ""
// @Failure 400 {object} map[string]string "bad json"
// @Failure 401 {object} map[string]string "bad user id"
// @Failure 404 {object} map[string]string "user not found"
// @Router /user/ [put]
func (s *Server) UpdateUserInfo(c *fiber.Ctx) error {
	var input UsernameInput
	err := c.BodyParser(&input)
	if err != nil || !user.CheckUsername(input.Username) {
		return ErrorResponse(c, http.StatusBadRequest, "bad json")
	}
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	err = s.services.UserService.UpdateUsername(userId, input.Username)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "user not found")
	}
	return c.Status(http.StatusNoContent).JSON(nil)
}

// UpdateProfilePic
// @Summary Обновить аватар пользователя
// @Description Обновляет аватар пользователя. Принимает изображение в формате GIF, JPG или PNG.
// @Security ApiAuth
// @Tags user
// @Accept multipart/form-data
// @Produce json
// @Param profile_pic formData file true "Изображение для аватара"
// @Success 204 {string} string ""
// @Failure 400 {object} map[string]string "bad form data"
// @Failure 400 {object} map[string]string "bad file format"
// @Failure 400 {object} map[string]string "unable to open file"
// @Failure 401 {object} map[string]string "bad user id"
// @Failure 404 {object} map[string]string "user not found"
// @Router /user/profile_pic [put]
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
		return ErrorResponse(c, http.StatusBadRequest, "unable to open file")
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "unable to open file")

	}
	suffix := filepath.Ext(ProfilePic.Filename)
	ValidFileSuffixForProfilePicture := []string{".gif", ".jpg", ".png"}
	if !slices.Contains(ValidFileSuffixForProfilePicture, suffix) {
		return ErrorResponse(c, http.StatusBadRequest, "bad file format")
	}
	user, err := s.services.UserService.GetUserById(userId)
	err = s.services.UserService.UpdateProfilePic(userId, fileBytes, user.Id.String(), suffix)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	return c.Status(http.StatusNoContent).JSON(nil)
}

// DailyReward
// @Summary Ежедневный вход
// @Description Получить награду за ежедневный вход
// @Security ApiAuth
// @Tags user
// @Produce json
// @Success 200 {object} user.DailyReward "Успех"
// @Failure 400 {object} map[string]string "next possible daily reward will available at {date}"
// @Failure 401 {object} map[string]string "bad user id"
// @Router /user/daily [post]
func (s *Server) DailyReward(c *fiber.Ctx) error {
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	reward, err := s.services.UserService.GetDaily(userId)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	return c.Status(http.StatusOK).JSON(reward)
}
