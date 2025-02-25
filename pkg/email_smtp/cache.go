package emailsmtp

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type IEmailCacheRepo interface {
	GetConfirmCode(email string) (string, time.Duration, error)
	SaveConfirmCode(email, code string) error
}

type EmailCacheRepo struct {
	db      *redis.Client
	CodeExp time.Duration
}

func NewEmailCacheRepo(db *redis.Client, CodeExp time.Duration) *EmailCacheRepo {
	return &EmailCacheRepo{db: db, CodeExp: CodeExp}
}

func (c *EmailCacheRepo) SaveConfirmCode(email, code string) error {
	ctx := context.Background()
	err := c.db.Set(ctx, email, code, c.CodeExp).Err()
	return err
}

func (c *EmailCacheRepo) GetConfirmCode(email string) (string, time.Duration, error) {
	var ttl time.Duration
	ctx := context.Background()
	code, err := c.db.Get(ctx, email).Result()
	if err != nil {
		return "", ttl, err
	}
	ttl, err = c.db.TTL(ctx, email).Result()

	return code, ttl, err
}
