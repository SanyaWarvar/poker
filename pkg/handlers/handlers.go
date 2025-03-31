package handlers

type Handler struct {
	services *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{services: s}
}
