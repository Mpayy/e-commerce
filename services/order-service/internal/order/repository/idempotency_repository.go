package repository

import (
	"context"
	"time"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_idempotency_repository.go
type IdempotencyRepository interface {
	Lock(ctx context.Context, idemKey string, lockTTL time.Duration) (bool, error)
	GetResponse(ctx context.Context, idemKey string) (*dto.IdempotencyResponse, error)
	SaveResponse(ctx context.Context, idemKey string, statusCode int, res *dto.OrderResponse, resultTTL time.Duration) error
	Unlock(ctx context.Context, idemKey string) error
}
