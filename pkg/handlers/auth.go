package handlers

import (
	"fmt"
	"net/http"
	"strings"

	_ "github.com/SanyaWarvar/poker/docs"
	"github.com/SanyaWarvar/poker/pkg/auth"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
func (h *Handler) SignUp(c *fiber.Ctx) error {
	var input1 UserInput

	err := c.BodyParser(&input1)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	input := user.User{Username: input1.Username, Email: input1.Email, Password: input1.Password}

	if valid := input.IsValid(); !valid {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid username or password")
	}

	err = h.services.UserService.CreateUser(input)
	if err != nil {
		fmt.Println(err.Error())
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

// SignInOutput
// @Schema
type SignInOutput struct {
	Tokens auth.RefreshInput `json:"tokens"`
	User   user.User         `json:"user"`
}

// SignIn
// @Summary Вход
// @Description Вход в аккаунт с подтвержденной почтой
// @Tags auth
// @Accept json
// @Produce json
// @Param body body EmailAndPasswordInput true "Данные пользователя"
// @Success 201 {object} SignInOutput "Успешный ответ"
// @Failure 400 {object} ErrorResponseStruct "Invalid json"
// @Failure 401 {object} ErrorResponseStruct "Invalid email or password"
// @Failure 403 {object} ErrorResponseStruct "Email not confirmed""
// @Failure 500 {object} ErrorResponseStruct "Failed to generate tokens"
// @Router /auth/sign_in [post]
func (h *Handler) SignIn(c *fiber.Ctx) error {
	var input EmailAndPasswordInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	user, err := h.services.UserService.GetUserByEP(input.Email, input.Password)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
	}
	user.GenerateUrl()
	if !user.IsEmailConfirmed {
		return ErrorResponse(c, http.StatusForbidden, "Email not confirmed")
	}

	accessToken, refreshToken, _, err := h.services.JwtService.GeneratePairToken(user.Id)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	return c.Status(http.StatusCreated).JSON(SignInOutput{
		Tokens: auth.RefreshInput{AccessToken: accessToken, RefreshToken: refreshToken},
		User:   user,
	})
}

// RefreshToken
// @Summary Обновление токенов
// @Description Обновляет хедеры с авторизационными токенами
// @Tags auth
// @Produce json
// @Param body body auth.RefreshInput true "Данные пользователя"
// @Success 201 {object} auth.RefreshInput "Успешный ответ"
// @Failure 400 {object} ErrorResponseStruct "Refresh token missing"
// @Failure 400 {object} ErrorResponseStruct "Access token missing"
// @Failure 400 {object} ErrorResponseStruct "Bad access token"
// @Failure 401 {object} ErrorResponseStruct "Bad refresh token"
// @Failure 500 {object} ErrorResponseStruct "Failed to generate tokens"
// @Router /auth/refresh_token [post]
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var input auth.RefreshInput

	err := c.BodyParser(&input)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "no access or refresh token")
	}

	accessToken, err := h.services.JwtService.ParseToken(input.AccessToken, false)
	fmt.Println(err)
	if err != nil && err != jwt.ErrTokenExpired {
		return ErrorResponse(c, http.StatusBadRequest, "bad access token")
	}

	isTokenValid := h.services.JwtService.CheckRefreshTokenExp(accessToken.RefreshId)

	if !isTokenValid {
		return ErrorResponse(c, http.StatusUnauthorized, "bad refresh token")
	}

	newAccessToken, newRefreshToken, _, err := h.services.JwtService.GeneratePairToken(accessToken.UserId)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to generate tokens")
	}

	return c.Status(http.StatusCreated).JSON(auth.RefreshInput{AccessToken: newAccessToken, RefreshToken: newRefreshToken})

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
func (h *Handler) SendCode(c *fiber.Ctx) error {
	var input EmailAndPasswordInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if err := h.services.EmailSmtpService.SendConfirmEmailMessage(input.Email); err != nil {
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
// @Failure 400 {object} ErrorResponseStruct "already confirmed"
// @Failure 400 {object} ErrorResponseStruct "Invalid confirmation code"
// @Failure 500 {object} ErrorResponseStruct "Failed to confirm code"
// @Router /auth/confirm_email [post]
func (h *Handler) ConfirmCode(c *fiber.Ctx) error {
	var input ConfirmCodeInput

	if err := c.BodyParser(&input); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid json")
	}

	if err := h.services.EmailSmtpService.ConfirmEmail(input.Email, input.Code); err != nil {
		if err.Error() == "Bad code" {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid confirmation code")
		} else if err.Error() == "already confirmed" {
			return ErrorResponse(c, http.StatusBadRequest, "already confirmed")
		}
		return ErrorResponse(c, http.StatusBadRequest, "Failed to confirm code")
	}
	c.Status(http.StatusCreated)
	return c.JSON(map[string]interface{}{
		"message": "Email confirmed",
	})
}
