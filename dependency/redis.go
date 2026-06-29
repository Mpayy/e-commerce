package dependency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const AuthPrefix = "auth:session:"

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../internal/mocks/mock_redis.go
type Redis interface {
	CheckToRedis(ctx context.Context, key string) (bool, error)
	SetToRedis(ctx context.Context, key string, value any, exp time.Duration) error
	DeleteFromRedis(ctx context.Context, key string) error
}

type RedisImpl struct {
	Client *redis.Client
}

func NewRedis(config *viper.Viper) Redis {
	addr := fmt.Sprintf("%s:%d", config.GetString("REDIS_HOST"), config.GetInt("REDIS_PORT"))
	password := config.GetString("REDIS_PASSWORD")
	db := config.GetInt("REDIS_DB")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisImpl{Client: client}
}

func (r *RedisImpl) CheckToRedis(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	response := result > 0

	return response, nil
}

func (r *RedisImpl) SetToRedis(ctx context.Context, key string, value any, exp time.Duration) error {
	err := r.Client.Set(ctx, key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisImpl) DeleteFromRedis(ctx context.Context, key string) error {
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
