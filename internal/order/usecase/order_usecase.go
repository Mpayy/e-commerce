package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/order/dto"
)

type OrderUsecase interface {
	Checkout(ctx context.Context, userID uint) (*dto.CheckoutResponse, error)
}