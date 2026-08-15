package repository

import (
	"context"
	"time"

	"github.com/Mpayy/e-commerce/monolith/internal/user/entity"
	"github.com/redis/go-redis/v9"
)

type UserRedisRepositoryImpl struct {
	Client *redis.Client
}

func NewUserRedisRepository(client *redis.Client) UserRedisRepository {
	return &UserRedisRepositoryImpl{Client: client}
}

func (r *UserRedisRepositoryImpl) SaveSession(ctx context.Context, token string, authData []byte, ttl time.Duration) error {
	err := r.Client.Set(ctx, entity.AuthPrefix+token, authData, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRedisRepositoryImpl) DeleteSession(ctx context.Context, token string) error {
	err := r.Client.Del(ctx, entity.AuthPrefix+token).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRedisRepositoryImpl) SessionExists(ctx context.Context, token string) (bool, error) {
	exists, err := r.Client.Exists(ctx, entity.AuthPrefix+token).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}
