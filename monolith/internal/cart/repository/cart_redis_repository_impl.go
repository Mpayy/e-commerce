package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/redis/go-redis/v9"
)

const CartPrefix = "cart:"

const CartTTL = 24 * time.Hour * 7

type CartRedisRepositoryImpl struct {
	rdb *redis.Client
}

func NewCartRedisRepository(rdb *redis.Client) CartRedisRepository {
	return &CartRedisRepositoryImpl{rdb: rdb}
}

func (r *CartRedisRepositoryImpl) AddItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))
	err := r.rdb.HIncrBy(ctx, CartPrefix+userIDStr, productIDStr, int64(quantity)).Err()
	if err != nil {
		return err
	}

	err = r.rdb.Expire(ctx, CartPrefix+userIDStr, CartTTL).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) UpdateItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))

	exists, err := r.rdb.HExists(ctx, CartPrefix+userIDStr, productIDStr).Result()
	if err != nil {
		return err
	}

	if !exists {
		return apperror.ErrRecordNotFound
	}

	err = r.rdb.HSet(ctx, CartPrefix+userIDStr, productIDStr, int64(quantity)).Err()
	if err != nil {
		return err
	}

	err = r.rdb.Expire(ctx, CartPrefix+userIDStr, CartTTL).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) RemoveItem(ctx context.Context, userID uint, productID uint) error {
	userIDStr := strconv.Itoa(int(userID))
	productIDStr := strconv.Itoa(int(productID))
	err := r.rdb.HDel(ctx, CartPrefix+userIDStr, productIDStr).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *CartRedisRepositoryImpl) GetCart(ctx context.Context, userID uint) (map[uint]int, error) {
	userIDStr := strconv.Itoa(int(userID))
	cart, err := r.rdb.HGetAll(ctx, CartPrefix+userIDStr).Result()
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
	err := r.rdb.Del(ctx, CartPrefix+userIDStr).Err()
	if err != nil {
		return err
	}

	return nil
}
