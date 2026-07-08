package cartrepository

import (
	"context"
	"strconv"

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
		return err
	}

	err = r.RedisClient.Expire(ctx, dependency.CartPrefix+userIDStr, dependency.CartTTL).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) UpdateItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))
	err := r.RedisClient.HSet(ctx, dependency.CartPrefix+userIDStr, productIDStr, int64(quantity)).Err()
	if err != nil {
		return err
	}

	err = r.RedisClient.Expire(ctx, dependency.CartPrefix+userIDStr, dependency.CartTTL).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) RemoveItem(ctx context.Context, userID uint, productID uint) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))
	err := r.RedisClient.HDel(ctx, dependency.CartPrefix+userIDStr, productIDStr).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) GetCart(ctx context.Context, userID uint) (map[uint]int, error) {
	userIDStr := strconv.Itoa(int(userID))
	cart, err := r.RedisClient.HGetAll(ctx, dependency.CartPrefix+userIDStr).Result()
	if err != nil {
		return nil, err
	}

	cartMap := make(map[uint]int)
	for productIDStr, quantityStr := range cart {
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			return nil, err
		}
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			return nil, err
		}
		cartMap[uint(productID)] = quantity
	}

	return cartMap, nil
}

func (r *CartRedisRepositoryImpl) ClearCart(ctx context.Context, userID uint) error {
	userIDStr := strconv.Itoa(int(userID))
	err := r.RedisClient.Del(ctx, dependency.CartPrefix+userIDStr).Err()
	if err != nil {
		return err
	}

	return nil
}