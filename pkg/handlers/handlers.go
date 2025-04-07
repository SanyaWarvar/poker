package handlers

import "github.com/SanyaWarvar/poker/pkg/game"

type Handler struct {
	services *Service
	engine   *game.HoldemEngine
}

func NewHandler(s *Service, e *game.HoldemEngine) *Handler {
	return &Handler{services: s, engine: e}
}
