package middleware

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const AuthPrefix = "auth:session:"

type RedisSessionChecker struct {
	rdb *redis.Client
}

func NewRedisSessionChecker(rdb *redis.Client) *RedisSessionChecker {
	return &RedisSessionChecker{rdb: rdb}
}

func (r *RedisSessionChecker) SessionExists(ctx context.Context, token string) (bool, error) {
	exists, err := r.rdb.Exists(ctx, AuthPrefix+token).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}