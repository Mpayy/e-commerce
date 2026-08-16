package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/monolith/internal/cart/dto"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_cart_usecase.go
type CartUsecase interface {
	AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error
	UpdateCartItem(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveFromCart(ctx context.Context, userID uint, productID uint) error
	GetCartDetail(ctx context.Context, userID uint) (*dto.CartDetailResponse, error)
}