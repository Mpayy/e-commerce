package cartusecase

import "context"

type CartUsecase interface {
	AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error
}