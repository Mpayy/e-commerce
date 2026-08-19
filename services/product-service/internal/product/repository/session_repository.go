package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const AuthPrefix = "auth:session:"

type SessionRepository interface {
	SessionExists(ctx context.Context, token string) (bool, error)
}

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return &sessionRepository{rdb: rdb}
}

func (r *sessionRepository) SessionExists(ctx context.Context, token string) (bool, error) {
	exists, err := r.rdb.Exists(ctx, AuthPrefix+token).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}
