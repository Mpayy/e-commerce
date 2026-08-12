package repository

import (
	"context"
	"time"
)

//go:generate mockery
//mockery:generate: true
//mockery:filename: ../mocks/mock_user_redis_repository.go
type UserRedisRepository interface {
	SaveSession(ctx context.Context, token string, authData []byte, ttl time.Duration) error
	DeleteSession(ctx context.Context, token string) error
	SessionExists(ctx context.Context, token string) (bool, error)
}
