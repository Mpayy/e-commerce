package cartrepository

import (
	"context"
	"strconv"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/dependency"
	"github.com/redis/go-redis/v9"
)

type CartRedisRepositoryImpl struct {
	RedisClient *redis.Client
}

func NewCartRedisRepository(client *redis.Client) CartRedisRepository {
	return &CartRedisRepositoryImpl{RedisClient: client}
}

func (r *CartRedisRepositoryImpl) AddItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))
	err := r.RedisClient.HIncrBy(ctx, dependency.CartPrefix+userIDStr, productIDStr, int64(quantity)).Err()
	if err != nil {
		return apperror.ErrInternalServer
	}

	err = r.RedisClient.Expire(ctx, dependency.CartPrefix+userIDStr, dependency.CartTTL).Err()
	if err != nil {
		return apperror.ErrInternalServer
	}

	return nil
}