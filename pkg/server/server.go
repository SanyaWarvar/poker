package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
)

type Server struct {
	services Service
}

func NewServer(s *Service) *Server {
	return &Server{services: *s}
}

func (s *Server) CreateApp() *fiber.App {
	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format:     "[${ip}:${port}] ${time} ${status} - ${method} ${path}\n",
		TimeFormat: "15:04:05 02-Jan-2006",
		TimeZone:   "Asia/Krasnoyarsk",
	}))

	auth := app.Group("/auth")
	{
		auth.Post("/sign_up", s.SignUp)
		auth.Post("/send_code", s.SendCode)
		auth.Post("/confirm_email", s.ConfirmCode)
		auth.Post("/sign_in", s.SignIn)
		auth.Post("/refresh_token", s.RefreshToken)
	}

	return app
}

func (s *Server) Run(port string) {
	app := s.CreateApp()

	logrus.Fatal(app.Listen(":" + port))
}
