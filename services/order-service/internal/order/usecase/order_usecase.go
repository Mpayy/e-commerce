package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
)

type OrderUsecase interface {
	Checkout(ctx context.Context, userID uint) (*dto.OrderResponse, error)
	GetOrderHistory(ctx context.Context, userID uint, filter *dto.OrderFilter) (*dto.OrderHistoryResponse, error)
	GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*dto.OrderResponse, error)
}
