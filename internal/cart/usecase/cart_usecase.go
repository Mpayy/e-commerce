package cartusecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/cart/dto"
)

type CartUsecase interface {
	AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error
	UpdateCartItem(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveFromCart(ctx context.Context, userID uint, productID uint) error
	GetCartDetail(ctx context.Context, userID uint) (*dto.CartDetailResponse, error)
}