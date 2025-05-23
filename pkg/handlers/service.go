package handlers

import (
	"github.com/SanyaWarvar/poker/pkg/auth"
	emailsmtp "github.com/SanyaWarvar/poker/pkg/email_smtp"
	"github.com/SanyaWarvar/poker/pkg/game"
	"github.com/SanyaWarvar/poker/pkg/notifications"
	"github.com/SanyaWarvar/poker/pkg/user"
)

type Service struct {
	JwtService          auth.IJwtManagerService
	UserService         user.IUserService
	EmailSmtpService    emailsmtp.IEmailSmtpService
	HoldemService       game.IHoldemService
	NotificationService notifications.INotificationService
}

func NewService(repos *Repository) *Service {
	return &Service{
		JwtService:          auth.NewJwtManagerService(repos.JwtRepo),
		UserService:         user.NewUserService(repos.UserRepo, repos.UserCacheRepo),
		EmailSmtpService:    emailsmtp.NewEmailSmtpService(repos.EmailSmtpRepo, repos.EmailSmtpCacheRepo),
		HoldemService:       game.NewHoldemService(repos.HoldemRepo, repos.UserRepo),
		NotificationService: notifications.NewNotificationService(repos.NotificationRepo),
	}
}
