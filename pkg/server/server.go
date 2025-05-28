package server

import (
	"net/http"

	_ "github.com/SanyaWarvar/poker/docs"
	"github.com/SanyaWarvar/poker/pkg/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer(h *handlers.Handler) *Server {
	return &Server{handler: h}
}

// @title Card House API
// @version 1.0
// @description This is a poker server
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host https://poker-tt7i.onrender.com
// @BasePath /
func (s *Server) CreateApp() *fiber.App {
	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format:     "[${ip}:${port}] ${time} ${status} - ${method} ${path}\n",
		TimeFormat: "15:04:05 02-Jan-2006",
		TimeZone:   "Asia/Krasnoyarsk",
	}))

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	app.Head("check_health", func(c *fiber.Ctx) error {
		c.Status(http.StatusOK)
		return c.JSON(map[string]string{"details": "ok"})
	})
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Static("/profiles", "./user_data/profile_pictures")

	auth := app.Group("/auth")
	{
		auth.Post("/sign_up", s.handler.SignUp)
		auth.Post("/send_code", s.handler.SendCode)
		auth.Post("/confirm_email", s.handler.ConfirmCode)
		auth.Post("/sign_in", s.handler.SignIn)
		auth.Post("/refresh_token", s.handler.RefreshToken)
	}

	user := app.Group("/user", s.handler.CheckAuthMiddleware)
	{
		user.Get(":username", s.handler.GetUser)
		user.Get("/byId/:id", s.handler.GetUserById)
		user.Put("/", s.handler.UpdateUserInfo)
		user.Put("/profile_pic", s.handler.UpdateProfilePic)
		user.Post("/daily", s.handler.DailyReward)
	}

	lobby := app.Group("/lobby", s.handler.CheckAuthMiddleware)
	{
		lobby.Get("/", s.handler.GetMyLobby)
		lobby.Get("/all/:page", s.handler.GetAllLobbies)
		lobby.Post("/", s.handler.CreateLobby)
	}
	{
		app.Get("ws/enter", websocket.New(s.handler.EnterInLobby))
		app.Get("ws/notifications", websocket.New(s.handler.NotificationsConnect))
	}

	return app
}

func (s *Server) Run(port string) {
	app := s.CreateApp()
	logrus.Fatal(app.Listen(":" + port))
}
