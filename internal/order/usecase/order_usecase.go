package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/order/dto"
)

type OrderUsecase interface {
	Checkout(ctx context.Context, userID uint) (*dto.OrderResponse, error)
	GetOrderHistory(ctx context.Context, userID uint) (*dto.OrderHistoryResponse, error)
	GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*dto.OrderResponse, error)
}