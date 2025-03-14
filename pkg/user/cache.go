package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IUserCacheRepo interface {
	GetLastDailyReward(userId uuid.UUID) (time.Time, error)
	SaveLastDailyReward(userId uuid.UUID) error
}

type UserCacheRepo struct {
	db *redis.Client
}

func NewUserCacheRepo(db *redis.Client) *UserCacheRepo {
	return &UserCacheRepo{db: db}
}

func (c *UserCacheRepo) SaveLastDailyReward(userId uuid.UUID) error {
	ctx := context.Background()
	err := c.db.Set(ctx, fmt.Sprintf("last_reward_%s", userId.String()), time.Now().Unix(), time.Hour*24).Err()
	return err
}

func (c *UserCacheRepo) GetLastDailyReward(userId uuid.UUID) (time.Time, error) {
	var outputTime time.Time
	ctx := context.Background()
	res := c.db.Get(ctx, fmt.Sprintf("last_reward_%s", userId.String()))
	t, err := res.Result()
	if err != nil {
		return outputTime, err
	}
	fmt.Println(t)
	timeUnix, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return outputTime, err
	}
	outputTime = time.Unix(timeUnix, 0)
	return outputTime, err
}
