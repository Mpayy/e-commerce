package repository

import (
	"context"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_cart_repository.go
type CartRedisRepository interface {
	AddItem(ctx context.Context, userID uint, productID uint, quantity int) error
	UpdateItem(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveItem(ctx context.Context, userID uint, productID uint) error
	GetCart(ctx context.Context, userID uint) (map[uint]int, error)
	ClearCart(ctx context.Context, userID uint) error
}
