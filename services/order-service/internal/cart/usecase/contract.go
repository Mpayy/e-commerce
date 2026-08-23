package usecase

import "context"


//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_cart_service.go
type CartService interface {
	GetRawCart(ctx context.Context, userID uint) (map[uint]int, error)
	ClearCart(ctx context.Context, userID uint) error
}
