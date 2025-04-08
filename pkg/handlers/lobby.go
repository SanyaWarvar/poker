package handlers

import (
	"net/http"
	"time"

	"github.com/SanyaWarvar/poker/pkg/game"
	"github.com/SanyaWarvar/poker/pkg/holdem"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetMyLobby
// @Summary Получить id лобби в котором находишься
// @Description Получить id лобби в котором находишься
// @Security ApiAuth
// @Tags lobby
// @Produce json
// @Success 200 {object} holdem.TableConfig "Успех"
// @Failure 400 {object} map[string]string "точно не знаю что тут может выпасть. наверное что то в духе lobby not found"
// @Failure 401 {object} map[string]string "bad user id"
// @Router /lobby/ [get]
func (h *Handler) GetMyLobby(c *fiber.Ctx) error {
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	lobby, err := h.services.HoldemService.GetLobbyByPId(userId)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error()) //TODO подумать хорошенько
	}
	return c.Status(http.StatusOK).JSON(lobby)
}

// GetAllLobbies
// @Summary Получить список лобби
// @Description Получить список лобби с пагинацией (размер страницы - 50)
// @Security ApiAuth
// @Tags lobby
// @Produce json
// @Param page query int true "Номер страницы" minimum(1)
// @Success 200 {object} []holdem.TableConfig "Список лобби"
// @Failure 400 {object} map[string]string "Неверный параметр страницы"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /lobby/all/{page} [get]
func (h *Handler) GetAllLobbies(c *fiber.Ctx) error {
	page, err := c.ParamsInt("page")
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad page param")
	}
	lobbies, err := h.services.HoldemService.GetLobbyList(page)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	return c.Status(http.StatusOK).JSON(lobbies)
}

// TableConfigInput
// @Schema
type TableConfigInput struct {
	BlindIncreaseTime string `json:"blind_increase_time" binding:"reqired" example:"15m"`
	MaxPlayers        int    `json:"max_players" binding:"reqired" example:"7"`
	EnterAfterStart   bool   `json:"cache_game" binding:"reqired" example:"true"` //true = cache game. false = sit n go
	SmallBlind        int    `json:"small_blind" binding:"reqired" example:"100"`
	Ante              int    `json:"ante" example:"25"`
	BankAmount        int    `json:"bank_amount"`
}

// CreateLobby
// @Summary Создать лобби
// @Description Создаить лобби
// @Security ApiAuth
// @Tags lobby
// @Produce json
// @Param body body TableConfigInput true "Данные для лобби"
// @Success 201 {object} string "id лобби"
// @Failure 400 {object} map[string]string "точно не знаю что тут может выпасть"
// @Failure 401 {object} map[string]string "bad user id"
// @Router /lobby/ [post]
func (h *Handler) CreateLobby(c *fiber.Ctx) error {
	var input TableConfigInput
	err := c.BodyParser(&input)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	minPlayers := 2
	if !input.EnterAfterStart {
		minPlayers = input.MaxPlayers
	}
	userIdInterface := c.Locals("userId")
	userId, ok := userIdInterface.(uuid.UUID)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "bad user id")
	}
	blindsIncreaseTime, err := time.ParseDuration(input.BlindIncreaseTime)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "bad user id")
	}
	cfg := holdem.NewTableConfig(
		blindsIncreaseTime,
		input.MaxPlayers,
		minPlayers,
		input.SmallBlind,
		input.Ante,
		input.BankAmount,
		input.EnterAfterStart,
		0,
	)

	lobbyId, err := h.services.HoldemService.CreateLobby(cfg, userId)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	h.services.HoldemService.AddObserver(lobbyId, h.engine.Observer)
	h.services.HoldemService.AddObserver(lobbyId, h.engine.Lt)
	h.engine.NewLobby(lobbyId, userId, game.LobbyInfo{
		GameStarted:  false,
		PlayersCount: 0,
		MinPlayers:   minPlayers,
		LastActivity: time.Now(),
		TTL:          game.DefaultTTL,
		TTS:          game.DefaultTTS,
	})

	return c.Status(http.StatusCreated).JSON(map[string]string{"lobby_id": lobbyId.String()})
}

// LobbyIdInput
// @Schema
type LobbyIdInput struct {
	LobbyId uuid.UUID `json:"lobby_id" binding:"required" example:"2854a298-61f5-468b-baa5-df4c273f2d06"`
}
