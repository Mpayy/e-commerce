package usecase

import "context"

type CartService interface {
	GetRawCart(ctx context.Context, userID uint) (map[uint]int, error)
	ClearCart(ctx context.Context, userID uint) error
}
