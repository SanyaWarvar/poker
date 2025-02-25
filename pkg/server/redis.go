package server

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisDb(opt *redis.Options) (*redis.Client, error) {
	ctx := context.Background()
	client := redis.NewClient(opt)
	err := client.Ping(ctx).Err()
	return client, err
}
