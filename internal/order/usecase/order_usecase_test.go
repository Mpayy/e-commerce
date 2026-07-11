package usecase

import (
	"context"
	"errors"
	"testing"

	"io"

	cartMock "github.com/Mpayy/e-commerce/internal/cart/mocks"
	configMock "github.com/Mpayy/e-commerce/internal/mocks"
	orderentity "github.com/Mpayy/e-commerce/internal/order/entity"
	repoMock "github.com/Mpayy/e-commerce/internal/order/mocks"
	productentity "github.com/Mpayy/e-commerce/internal/product/entity"
	productMock "github.com/Mpayy/e-commerce/internal/product/mocks"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func setupOrderUsecase(t *testing.T) (OrderUsecase, *productMock.MockProductService, *cartMock.MockCartService, *repoMock.MockOrderRepository, *configMock.MockTransaction) {
	orderRepository := repoMock.NewMockOrderRepository(t)
	cartService := cartMock.NewMockCartService(t)
	productService := productMock.NewMockProductService(t)
	transactionMock := configMock.NewMockTransaction(t)
	log := newTestLogger()

	orderUsecase := NewOrderUsecase(orderRepository, transactionMock, log, cartService, productService)
	return orderUsecase, productService, cartService, orderRepository, transactionMock
}

//Before run test always run this command -> go clean -testcache

// go test -v ./internal/order/usecase -run "TestOrderUsecase_Checkout"
func TestOrderUsecase_Checkout(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_Checkout/success_checkout"
	t.Run("success_checkout", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{
			1: 10,
			2: 5,
		}

		products := []productentity.Product{
			{
				ID:    1,
				Name:  "Produk 1",
				Price: 10000,
				Stock: 10,
			},
			{
				ID:    2,
				Name:  "Produk 2",
				Price: 20000,
				Stock: 5,
			},
		}

		cartService.On("GetRawCart", mock.Anything, userID).
			Return(rawCart, nil)

		productService.On("GetProductsByIDs", mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.On("DecreaseStock", mock.Anything, uint(1), 10).
			Return(nil)
		productService.On("DecreaseStock", mock.Anything, uint(2), 5).
			Return(nil)

		orderRepository.On("CreateOrderWithItems", mock.Anything, mock.Anything, mock.Anything).
			Return(func(ctx context.Context, order *orderentity.Order, items []orderentity.OrderItem) error {
				order.ID = 6
				order.InvoiceNumber = "INV-20260710-000006"
				return nil
			})

		cartService.On("ClearCart", mock.Anything, userID).
			Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, uint(6), result.OrderID)
		assert.Equal(t, "INV-20260710-000006", result.InvoiceNumber)
		assert.Equal(t, "PAID", result.Status)
		assert.Equal(t, float64(200000), result.TotalAmount)
		assert.Len(t, result.Items, 2)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_Checkout/failed_cart_empty"
	t.Run("failed_cart_empty", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{}

		cartService.On("GetRawCart", mock.Anything, userID).Return(rawCart, nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCartEmpty)

		productService.AssertNotCalled(t, "GetProductsByIDs", mock.Anything, mock.Anything)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_Checkout/failed_insufficient_stock_in_mid_of_loop"
	t.Run("failed_insufficient_stock_in_mid_of_loop", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)
		userID := uint(1)

		rawCart := map[uint]int{
			1: 10,
			2: 5,
		}

		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 15},
			{ID: 2, Name: "Produk 2", Price: 20000, Stock: 2},
		}

		cartService.On("GetRawCart", mock.Anything, userID).Return(rawCart, nil)
		productService.On("GetProductsByIDs", mock.Anything, mock.Anything).Return(products, nil)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.On("DecreaseStock", mock.Anything, uint(1), int(10)).Return(nil)
		productService.On("DecreaseStock", mock.Anything, uint(2), int(5)).Return(apperror.ErrInsufficientStock)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInsufficientStock)

		orderRepository.AssertNotCalled(t, "CreateOrderWithItems", mock.Anything, mock.Anything, mock.Anything)
		cartService.AssertNotCalled(t, "ClearCart", mock.Anything, mock.Anything)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_Checkout/failed_product_not_found_mid_loop"
	t.Run("failed_product_not_found_mid_loop", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 10}

		cartService.On("GetRawCart", mock.Anything, userID).Return(rawCart, nil)

		productService.On("GetProductsByIDs", mock.Anything, mock.Anything).
			Return([]productentity.Product{}, nil)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)

		productService.AssertNotCalled(t, "DecreaseStock", mock.Anything, uint(1), 10)
		orderRepository.AssertNotCalled(t, "CreateOrderWithItems", mock.Anything, mock.Anything, mock.Anything)
		cartService.AssertNotCalled(t, "ClearCart", mock.Anything, mock.Anything)
	})
}

// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderHistory"
func TestOrderUsecase_GetOrderHistory(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderHistory/success_get_history_with_items"
	t.Run("success_get_history_with_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrders := []orderentity.Order{
			{ID: 101, UserID: userID, InvoiceNumber: "INV-001", TotalAmount: 50000, Status: "PAID"},
		}
		mockItems := []orderentity.OrderItem{
			{OrderID: 101, ProductID: 1, ProductName: "Sepatu", Price: 25000, Quantity: 2, Subtotal: 50000},
		}

		orderRepository.On("FindByUserID", mock.Anything, userID).Return(mockOrders, mockItems, nil)

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Orders, 1)
		assert.Equal(t, uint(101), result.Orders[0].OrderID)
		assert.Len(t, result.Orders[0].Items, 1)
		assert.Equal(t, "Sepatu", result.Orders[0].Items[0].ProductName)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderHistory/success_get_history_empty"
	t.Run("success_get_history_empty", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.On("FindByUserID", mock.Anything, userID).Return([]orderentity.Order{}, []orderentity.OrderItem{}, nil)

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Orders, 0) // Harus aman berupa array kosong
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderHistory/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.On("FindByUserID", mock.Anything, userID).Return(nil, nil, errors.New("unexpexted error"))

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail"
func TestOrderUsecase_GetOrderDetail(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	orderID := uint(101)

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/success_get_detail"
	t.Run("success_get_detail", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &orderentity.Order{ID: orderID, UserID: userID, InvoiceNumber: "INV-001", TotalAmount: 30000, Status: "PAID"}
		mockItems := []orderentity.OrderItem{
			{OrderID: orderID, ProductID: 5, ProductName: "Kopi", Price: 15000, Quantity: 2, Subtotal: 30000},
		}

		orderRepository.On("FindByID", mock.Anything, orderID).Return(mockOrder, nil)
		orderRepository.On("FindItemsByOrderID", mock.Anything, orderID).Return(mockItems, nil)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "INV-001", result.InvoiceNumber)
		assert.Len(t, result.Items, 1)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_order_not_found"
	t.Run("failed_order_not_found", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.On("FindByID", mock.Anything, orderID).Return(nil, apperror.ErrNotFound)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_wrong_ownership"
	t.Run("failed_wrong_ownership", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrderWithWrongOwner := &orderentity.Order{
			ID:            orderID,
			UserID:        uint(99),
			InvoiceNumber: "INV-HACKER",
		}

		orderRepository.On("FindByID", mock.Anything, orderID).Return(mockOrderWithWrongOwner, nil)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)

		orderRepository.AssertNotCalled(t, "FindItemsByOrderID", mock.Anything, mock.Anything)
	})

	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_unexpected_db_error_on_items"
	t.Run("failed_unexpected_db_error_on_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &orderentity.Order{ID: orderID, UserID: userID}

		orderRepository.On("FindByID", mock.Anything, orderID).Return(mockOrder, nil)
		orderRepository.On("FindItemsByOrderID", mock.Anything, orderID).Return(nil, errors.New("table items locked"))

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}
