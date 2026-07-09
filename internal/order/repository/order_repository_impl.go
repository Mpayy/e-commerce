package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mpayy/e-commerce/internal/order/entity"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"gorm.io/gorm"
)

type OrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{DB: db}
}

func (r *OrderRepositoryImpl) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.GetTxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return r.DB.WithContext(ctx)
}

func (r *OrderRepositoryImpl) CreateOrderWithItems(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
	err := r.GetTx(ctx).Create(order).Error
	if err != nil {
		return err
	}

	invoiceNumber := fmt.Sprintf("INV-%s-%06d",
		order.CreatedAt.Format("20060102"),
		order.ID,
	)

	err = r.GetTx(ctx).Model(order).Update("invoice_number", invoiceNumber).Error
	if err != nil {
		return err
	}

	order.InvoiceNumber = invoiceNumber

	for i := range items {
		items[i].OrderID = order.ID
	}

	err = r.GetTx(ctx).Create(&items).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]entity.Order, []entity.OrderItem, error) {
	var orders []entity.Order
	var items []entity.OrderItem

	err := r.GetTx(ctx).Where("user_id = ?", userID).Find(&orders).Error
	if err != nil {
		return nil, nil, err
	}

	if len(orders) == 0 {
		return orders, items, nil
	}

	var orderIDs []uint
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	err = r.GetTx(ctx).Where("order_id IN ?", orderIDs).Find(&items).Error
	if err != nil {
		return nil, nil, err
	}
	return orders, items, nil
}

func (r *OrderRepositoryImpl) FindByID(ctx context.Context, orderID uint) (*entity.Order, error) {
	var order entity.Order
	err := r.GetTx(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) FindItemsByOrderID(ctx context.Context, orderID uint) ([]entity.OrderItem, error) {
	var items []entity.OrderItem
	err := r.GetTx(ctx).Where("order_id = ?", orderID).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
