package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/order/entity"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_order_repository.go
type OrderRepository interface {
	CreateOrderWithItems(ctx context.Context, order *entity.Order, items []entity.OrderItem) error
	FindByUserID(ctx context.Context, userID uint) ([]entity.Order, []entity.OrderItem, error)
	FindByID(ctx context.Context, orderID uint) (*entity.Order, error)
	FindItemsByOrderID(ctx context.Context, orderID uint) ([]entity.OrderItem, error)
}
