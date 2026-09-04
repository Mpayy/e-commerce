package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
)

type OrderUsecase interface {
	Checkout(ctx context.Context, userID uint) (*dto.OrderResponse, error)
	GetOrderHistory(ctx context.Context, userID uint, filter *dto.OrderFilter) (*dto.OrderHistoryResponse, error)
	GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*dto.OrderResponse, error)
	GetSalesAnalytics(ctx context.Context, req *dto.SalesAnalyticsRequest) (*dto.SalesAnalyticsResponse, error)
	GetAdminOrderList(ctx context.Context, filter *dto.AdminOrderListRequest) (*dto.AdminOrderListResponse, error)
	GetAdminOrderDetail(ctx context.Context, orderID uint) (*dto.OrderResponse, error)
}
