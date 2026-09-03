package repository

import (
	"context"
	"time"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/entity"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_order_repository.go
type OrderRepository interface {
	CreateOrderWithItems(ctx context.Context, order *entity.Order, items []entity.OrderItem) error
	FindByUserID(ctx context.Context, userID uint, page int, limit int) ([]entity.Order, int64, error)
	FindByID(ctx context.Context, orderID uint) (*entity.Order, error)
	GetDailyRevenueReport(ctx context.Context, from, to time.Time) ([]entity.DailyRevenueRow, error)
	GetTopProducts(ctx context.Context, from, to time.Time, limit int32) ([]entity.TopProductRow, error)
	GetAdminOrderList(ctx context.Context, filter *entity.OrderFilter) ([]entity.Order, error)
	CountAdminOrderList(ctx context.Context, filter *entity.OrderFilter) (int64, error)
}
