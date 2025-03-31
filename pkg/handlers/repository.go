package handlers

import (
	"github.com/SanyaWarvar/poker/pkg/auth"
	emailsmtp "github.com/SanyaWarvar/poker/pkg/email_smtp"
	"github.com/SanyaWarvar/poker/pkg/game"
	"github.com/SanyaWarvar/poker/pkg/user"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	JwtRepo            auth.IJwtManagerRepo
	UserRepo           user.IUserRepo
	UserCacheRepo      user.IUserCacheRepo
	EmailSmtpRepo      emailsmtp.IEmailSmtpRepo
	EmailSmtpCacheRepo emailsmtp.IEmailCacheRepo
	HoldemRepo         game.IHoldemRepo
}

func NewRepository(
	db *sqlx.DB,
	cacheDb *redis.Client,
	emailCfg *emailsmtp.EmailCfg,
	jwtCfg *auth.JwtManagerCfg,
) *Repository {

	return &Repository{
		JwtRepo:            auth.NewJwtManagerPostgres(db, jwtCfg),
		UserRepo:           user.NewUserPostgres(db),
		UserCacheRepo:      user.NewUserCacheRepo(cacheDb),
		EmailSmtpRepo:      emailsmtp.NewEmailSmtpPostgres(db, emailCfg),
		EmailSmtpCacheRepo: emailsmtp.NewEmailCacheRepo(cacheDb, emailCfg.CodeExp),
		HoldemRepo:         game.NewHoldemRepo(),
	}
}
