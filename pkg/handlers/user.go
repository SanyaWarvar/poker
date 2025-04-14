package handlers

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	_ "github.com/SanyaWarvar/poker/docs"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
func (h *Handler) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")

	if username == "" {
		return ErrorResponse(c, http.StatusBadRequest, "username cant be empty")
	}
	user, err := h.services.UserService.GetUserByUsername(username)
	if err != nil { // TODO возможно могут быть другие проблемы?
		return ErrorResponse(c, http.StatusNotFound, "user not found")
	}
	user.GenerateUrl()
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
func (h *Handler) UpdateUserInfo(c *fiber.Ctx) error {
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
	err = h.services.UserService.UpdateUsername(userId, input.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}

		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return ErrorResponse(c, http.StatusConflict, "username already taken")
		}

		return ErrorResponse(c, http.StatusInternalServerError, "failed to update username")
	}
	return c.Status(http.StatusNoContent).JSON(nil)
}

// ProfilePicUrlStruct
// @Schema
type ProfilePicUrlStruct struct {
	ProfilePicUrl string `json:"pic_url" example:"host/profiles/example.jpg"`
}

// UpdateProfilePic
// @Summary Обновить аватар пользователя
// @Description Обновляет аватар пользователя. Принимает изображение в формате GIF, JPG или PNG.
// @Security ApiAuth
// @Tags user
// @Accept multipart/form-data
// @Produce json
// @Param profile_pic formData file true "Изображение для аватара"
// @Success 200 {object} ProfilePicUrlStruct "Успешное обновление"
// @Failure 400 {object} map[string]string "bad form data"
// @Failure 400 {object} map[string]string "bad file format"
// @Failure 400 {object} map[string]string "unable to open file"
// @Failure 401 {object} map[string]string "bad user id"
// @Failure 404 {object} map[string]string "user not found"
// @Router /user/profile_pic [put]
func (h *Handler) UpdateProfilePic(c *fiber.Ctx) error {
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
	err = h.services.UserService.UpdateProfilePic(userId, fileBytes, suffix)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	user, _ := h.services.UserService.GetUserById(userId)
	user.GenerateUrl()
	return c.Status(http.StatusOK).JSON(ProfilePicUrlStruct{ProfilePicUrl: user.ProfilePicUrl})
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
func (h *Handler) DailyReward(c *fiber.Ctx) error {
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	reward, err := h.services.UserService.GetDaily(userId)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	return c.Status(http.StatusOK).JSON(reward)
}
