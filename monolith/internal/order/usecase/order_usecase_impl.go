package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	cartUC "github.com/Mpayy/e-commerce/monolith/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/monolith/internal/order/dto"
	"github.com/Mpayy/e-commerce/monolith/internal/order/entity"
	"github.com/Mpayy/e-commerce/monolith/internal/order/repository"
	productentity "github.com/Mpayy/e-commerce/monolith/internal/product/entity"
	productUC "github.com/Mpayy/e-commerce/monolith/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/google/uuid"
)

type OrderUsecaseImpl struct {
	orderRepository repository.OrderRepository
	transaction     transaction.Transaction
	log             *logger.Logger
	cartService     cartUC.CartService
	productService  productUC.ProductService
}

func NewOrderUsecase(orderRepository repository.OrderRepository, transaction transaction.Transaction, log *logger.Logger, cartService cartUC.CartService, productService productUC.ProductService) OrderUsecase {
	return &OrderUsecaseImpl{
		orderRepository: orderRepository,
		transaction:     transaction,
		log:             log,
		cartService:     cartService,
		productService:  productService,
	}
}

func (u *OrderUsecaseImpl) Checkout(ctx context.Context, userID uint) (*dto.OrderResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Attempting to checkout")

	rawCart, err := u.cartService.GetRawCart(ctx, userID)
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

	products, err := u.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[uint]productentity.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	var finalizedOrder entity.Order
	var finalizedOrderItems []entity.OrderItem
	checkoutID := uuid.NewString()

	err = u.transaction.WithTransaction(ctx, func(ctx context.Context) error {
		var orderItems []entity.OrderItem
		var grandTotal float64
		var decrements []productentity.BulkDecreaseStock

		for productID, qty := range rawCart {
			if qty <= 0 {
				continue
			}

			product, exists := productMap[productID]
			if !exists {
				return apperror.ErrProductNotFound
			}

			decrements = append(decrements, productentity.BulkDecreaseStock{
				ProductID: product.ID,
				Quantity:  qty,
			})

			subtotal := product.Price * float64(qty)
			grandTotal += subtotal

			orderItems = append(orderItems, entity.OrderItem{
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

		if err = u.productService.BulkDecreaseStock(ctx, checkoutID, decrements); err != nil {
			return err
		}

		order := entity.Order{
			UserID:      userID,
			TotalAmount: grandTotal,
			Status:      "PAID",
		}

		if err = u.orderRepository.CreateOrderWithItems(ctx, &order, orderItems); err != nil {
			restoreCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if restoreErr := u.productService.BulkRestoreStock(restoreCtx, checkoutID); restoreErr != nil {
				log.WithFields(logger.Fields{
					"checkout_id":    checkoutID,
					"restore_error":  restoreErr,
					"original_error": err,
				}).Error("CRITICAL: stock compensation failed after order creation failed — manual reconciliation required.")
			}
			return fmt.Errorf("failed to create order with items: %w", err)
		}

		finalizedOrder = order
		finalizedOrderItems = orderItems

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = u.cartService.ClearCart(ctx, userID)
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

	log.WithFields(logger.Fields{
		"order_id":     finalizedOrder.ID,
		"total_amount": finalizedOrder.TotalAmount,
		"items":        len(responseItems),
	}).Info("Checkout successful")

	return &dto.OrderResponse{
		OrderID:       finalizedOrder.ID,
		InvoiceNumber: finalizedOrder.InvoiceNumber,
		TotalAmount:   finalizedOrder.TotalAmount,
		Status:        finalizedOrder.Status,
		Items:         responseItems,
	}, nil
}

func (u *OrderUsecaseImpl) GetOrderHistory(ctx context.Context, userID uint) (*dto.OrderHistoryResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Getting order history")

	orders, items, err := u.orderRepository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
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

	log.WithFields(logger.Fields{
		"user_id":     userID,
		"order_count": len(responseOrders),
	}).Debug("Order history retrieved successfully")

	return &dto.OrderHistoryResponse{
		Orders: responseOrders,
	}, nil
}

func (u *OrderUsecaseImpl) GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*dto.OrderResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id":  userID,
		"order_id": orderID,
	})
	log.Debug("Getting order detail")

	order, err := u.orderRepository.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order detail: %w", err)
	}

	if order.UserID != userID {
		return nil, apperror.ErrOrderNotFound
	}

	items, err := u.orderRepository.FindItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order detail: %w", err)
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

	log.Debug("Order detail retrieved successfully")

	return &dto.OrderResponse{
		OrderID:       order.ID,
		InvoiceNumber: order.InvoiceNumber,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		Items:         responseItems,
	}, nil
}
