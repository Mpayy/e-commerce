package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/entity"
)

type UserRedisRepositoryImpl struct {
	client *redis.Client
}

func NewUserRedisRepository(client *redis.Client) UserRedisRepository {
	return &UserRedisRepositoryImpl{client: client}
}

func (r *UserRedisRepositoryImpl) SaveSession(ctx context.Context, token string, authData []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, entity.AuthPrefix+token, authData, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRedisRepositoryImpl) DeleteSession(ctx context.Context, token string) error {
	err := r.client.Del(ctx, entity.AuthPrefix+token).Err()
	if err != nil {
		return err
	}
	return nil
}
