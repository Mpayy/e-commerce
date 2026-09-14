package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/redis/go-redis/v9"
)

const idemKeyPrefix = "idempotency:key:"

type IdempotencyRepositoryImpl struct {
	client *redis.Client
}

func NewIdempotencyRepository(client *redis.Client) IdempotencyRepository {
	return &IdempotencyRepositoryImpl{
		client: client,
	}
}

func (i *IdempotencyRepositoryImpl) Lock(ctx context.Context, idemKey string, lockTTL time.Duration) (bool, error) {
	return i.client.SetNX(ctx, idemKeyPrefix+idemKey, "PROCESSING", lockTTL).Result()
}

func (i *IdempotencyRepositoryImpl) GetResponse(ctx context.Context, idemKey string) (*dto.IdempotencyResponse, error) {
	val, err := i.client.Get(ctx, idemKeyPrefix+idemKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if val == "PROCESSING" {
		return nil, nil
	}

	var res dto.IdempotencyResponse
	err = json.Unmarshal([]byte(val), &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (i *IdempotencyRepositoryImpl) SaveResponse(ctx context.Context, idemKey string, statusCode int, res *dto.OrderResponse, resultTTL time.Duration) error {
	idemRes := dto.IdempotencyResponse{
		StatusCode: statusCode,
		Body:       res,
	}

	bytes, err := json.Marshal(idemRes)
	if err != nil {
		return err
	}

	return i.client.Set(ctx, idemKeyPrefix+idemKey, bytes, resultTTL).Err()
}

func (i *IdempotencyRepositoryImpl) Unlock(ctx context.Context, idemKey string) error {
	return i.client.Del(ctx, idemKeyPrefix+idemKey).Err()
}
