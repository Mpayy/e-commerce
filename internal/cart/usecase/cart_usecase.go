package cartusecase

import "context"

type CartUsecase interface {
	AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error
	UpdateCartItem(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveFromCart(ctx context.Context, userID uint, productID uint) error
}