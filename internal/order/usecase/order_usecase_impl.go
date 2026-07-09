package usecase

import (
	"context"
	"errors"

	cartusecase "github.com/Mpayy/e-commerce/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/internal/order/dto"
	orderentity "github.com/Mpayy/e-commerce/internal/order/entity"
	"github.com/Mpayy/e-commerce/internal/order/repository"
	productentity "github.com/Mpayy/e-commerce/internal/product/entity"
	productusecase "github.com/Mpayy/e-commerce/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/sirupsen/logrus"
)

type OrderUsecaseImpl struct {
	OrderRepository repository.OrderRepository
	Trasaction      transaction.Transaction
	Log             *logrus.Logger
	CartService     cartusecase.CartService
	ProductService  productusecase.ProductService
}

func NewOrderUsecase(orderRepository repository.OrderRepository, trasaction transaction.Transaction, log *logrus.Logger, cartService cartusecase.CartService, productService productusecase.ProductService) OrderUsecase {
	return &OrderUsecaseImpl{
		OrderRepository: orderRepository,
		Trasaction:      trasaction,
		Log:             log,
		CartService:     cartService,
		ProductService:  productService,
	}
}

func (u *OrderUsecaseImpl) Checkout(ctx context.Context, userID uint) (*dto.OrderResponse, error) {
	u.Log.WithField("user_id", userID).Debug("Attempting to checkout")

	rawCart, err := u.CartService.GetRawCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(rawCart) == 0 {
		return nil, apperror.ErrCartEmpty
	}

	var productIDs []uint
	for productID := range rawCart {
		productIDs = append(productIDs, productID)
	}

	products, err := u.ProductService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[uint]productentity.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	var finalizedOrder orderentity.Order
	var finalizedOrderItems []orderentity.OrderItem

	err = u.Trasaction.WithTransaction(ctx, func(ctx context.Context) error {
		var orderItems []orderentity.OrderItem
		var grandTotal float64

		for productID, qty := range rawCart {
			if qty <= 0 {
				continue
			}

			product, exists := productMap[productID]
			if !exists {
				return apperror.ErrProductNotFound
			}

			if qty > product.Stock {
				return apperror.ErrInsufficientStock
			}

			err := u.ProductService.DecreaseStock(ctx, product.ID, qty)
			if err != nil {
				return err
			}

			subtotal := product.Price * float64(qty)
			grandTotal += subtotal

			orderItems = append(orderItems, orderentity.OrderItem{
				ProductID:   product.ID,
				ProductName: product.Name,
				Quantity:    qty,
				Subtotal:    subtotal,
				Price:       product.Price,
			})
		}

		if len(orderItems) == 0 {
			return apperror.ErrCartEmpty
		}

		order := orderentity.Order{
			UserID:      userID,
			TotalAmount: grandTotal,
			Status:      "PENDING",
		}

		err := u.OrderRepository.CreateOrderWithItems(ctx, &order, orderItems)
		if err != nil {
			u.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err,
			}).Error("Failed to create order with items")
			return apperror.ErrInternalServer
		}

		finalizedOrder = order
		finalizedOrderItems = orderItems

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = u.CartService.ClearCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responseItems []dto.OrderItemResponse
	for _, item := range finalizedOrderItems {
		product := productMap[item.ProductID]
		subtotal := item.Price * float64(item.Quantity)

		responseItems = append(responseItems, dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":      userID,
		"order_id":     finalizedOrder.ID,
		"total_amount": finalizedOrder.TotalAmount,
		"items":        len(responseItems),
	}).Debug("Checkout successful")

	return &dto.OrderResponse{
		OrderID:       finalizedOrder.ID,
		InvoiceNumber: finalizedOrder.InvoiceNumber,
		TotalAmount:   finalizedOrder.TotalAmount,
		Status:        finalizedOrder.Status,
		Items:         responseItems,
	}, nil
}

func (u *OrderUsecaseImpl) GetOrderHistory(ctx context.Context, userID uint) (*dto.OrderHistoryResponse, error) {
	orders, items, err := u.OrderRepository.FindByUserID(ctx, userID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get order history")
		return nil, apperror.ErrInternalServer
	}

	orderMap := make(map[uint][]dto.OrderItemResponse)
	for _, item := range items {
		orderMap[item.OrderID] = append(orderMap[item.OrderID], dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	responseOrders := []dto.OrderResponse{}
	for _, order := range orders {
		itemForThisOrder := orderMap[order.ID]

		if itemForThisOrder == nil {
			itemForThisOrder = []dto.OrderItemResponse{}
		}

		responseOrders = append(responseOrders, dto.OrderResponse{
			OrderID:       order.ID,
			InvoiceNumber: order.InvoiceNumber,
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			Items:         itemForThisOrder,
		})
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":     userID,
		"order_count": len(responseOrders),
	}).Debug("Order history retrieved successfully")

	return &dto.OrderHistoryResponse{
		Orders: responseOrders,
	}, nil
}

func (u *OrderUsecaseImpl) GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*dto.OrderResponse, error) {
	order, err := u.OrderRepository.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrOrderNotFound
		}
		u.Log.WithFields(logrus.Fields{
			"user_id":  userID,
			"order_id": orderID,
			"error":    err,
		}).Error("Failed to get order detail")
		return nil, apperror.ErrInternalServer
	}

	if order.UserID != userID {
		return nil, apperror.ErrOrderNotFound
	}

	items, err := u.OrderRepository.FindItemsByOrderID(ctx, orderID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id":  userID,
			"order_id": orderID,
			"error":    err,
		}).Error("Failed to get order detail")
		return nil, apperror.ErrInternalServer
	}

	responseItems := []dto.OrderItemResponse{}
	for _, item := range items {
		responseItems = append(responseItems, dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":  userID,
		"order_id": orderID,
	}).Debug("Order detail retrieved successfully")

	return &dto.OrderResponse{
		OrderID:       order.ID,
		InvoiceNumber: order.InvoiceNumber,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		Items:         responseItems,
	}, nil
}
