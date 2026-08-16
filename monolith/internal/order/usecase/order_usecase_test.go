package usecase

import (
	"context"
	"errors"
	"testing"

	"io"

	cartMock "github.com/Mpayy/e-commerce/monolith/internal/cart/mocks"
	"github.com/Mpayy/e-commerce/monolith/internal/order/entity"
	repoMock "github.com/Mpayy/e-commerce/monolith/internal/order/mocks"
	productentity "github.com/Mpayy/e-commerce/monolith/internal/product/entity"
	productMock "github.com/Mpayy/e-commerce/monolith/internal/product/mocks"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupOrderUsecase(t *testing.T) (OrderUsecase, *productMock.MockProductService, *cartMock.MockCartService, *repoMock.MockOrderRepository, *repoMock.MockTransaction) {
	orderRepository := repoMock.NewMockOrderRepository(t)
	cartService := cartMock.NewMockCartService(t)
	productService := productMock.NewMockProductService(t)
	transactionMock := repoMock.NewMockTransaction(t)
	log := newTestLogger()

	orderUsecase := NewOrderUsecase(orderRepository, transactionMock, log, cartService, productService)
	return orderUsecase, productService, cartService, orderRepository, transactionMock
}

func TestOrderUsecase_Checkout(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	dbErr := errors.New("unexpected database error")

	// 1. Success Checkout
	t.Run("success_checkout", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{
			1: 10,
			2: 5,
		}

		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 15},
			{ID: 2, Name: "Produk 2", Price: 20000, Stock: 10},
		}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.MatchedBy(func(ids []uint) bool {
				return len(ids) == 2
			})).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.EXPECT().
			BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
				order.ID = 6
				order.InvoiceNumber = "INV-20260710-000006"
				return nil
			})

		cartService.EXPECT().
			ClearCart(mock.Anything, userID).
			Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(6), result.OrderID)
		assert.Equal(t, "INV-20260710-000006", result.InvoiceNumber)
		assert.Equal(t, "PAID", result.Status)
		assert.Equal(t, float64(200000), result.TotalAmount) // (10000*10) + (20000*5)
		assert.Len(t, result.Items, 2)
	})

	// 2. Failed: GetRawCart Error
	t.Run("failed_get_raw_cart_error", func(t *testing.T) {
		usecase, _, cartService, _, _ := setupOrderUsecase(t)

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(nil, dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	// 3. Failed: Cart Initially Empty
	t.Run("failed_cart_empty_initial", func(t *testing.T) {
		usecase, _, cartService, _, _ := setupOrderUsecase(t)

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(map[uint]int{}, nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCartEmpty)
	})

	// 4. Failed: GetProductsByIDs Error
	t.Run("failed_get_products_by_ids_error", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(nil, dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	// 5. Failed: Cart Items Qty <= 0 Resulting in Empty Order Items
	t.Run("failed_cart_items_quantity_zero_or_negative", func(t *testing.T) {
		usecase, productService, cartService, _, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 0, 2: -1}
		products := []productentity.Product{
			{ID: 1, Name: "P1", Price: 10000},
			{ID: 2, Name: "P2", Price: 20000},
		}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCartEmpty)
	})

	// 6. Failed: Product Not Found in Map
	t.Run("failed_product_not_found_in_map", func(t *testing.T) {
		usecase, productService, cartService, _, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return([]productentity.Product{}, nil) // Returns empty product slice

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// 7. Failed: BulkDecreaseStock Error
	t.Run("failed_bulk_decrease_stock_error", func(t *testing.T) {
		usecase, productService, cartService, _, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{{ID: 1, Name: "P1", Price: 10000}}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.EXPECT().
			BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).
			Return(apperror.ErrInsufficientStock)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInsufficientStock)
	})

	// 8. Failed: CreateOrderWithItems Error & Compensation (BulkRestoreStock) Success
	t.Run("failed_create_order_with_items_and_restore_stock_success", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{{ID: 1, Name: "P1", Price: 10000}}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.EXPECT().
			BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			Return(dbErr)

		// Verification that compensation (BulkRestoreStock) is executed and succeeds
		productService.EXPECT().
			BulkRestoreStock(mock.Anything, mock.Anything).
			Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	// 9. Failed: CreateOrderWithItems Error & Compensation (BulkRestoreStock) Fails
	t.Run("failed_create_order_with_items_and_restore_stock_fails", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{{ID: 1, Name: "P1", Price: 10000}}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.EXPECT().
			BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			Return(dbErr)

		// Compensation fails (triggers CRITICAL log path)
		productService.EXPECT().
			BulkRestoreStock(mock.Anything, mock.Anything).
			Return(dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	// 10. Failed: Transaction Failure
	t.Run("failed_transaction_execution", func(t *testing.T) {
		usecase, productService, cartService, _, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{{ID: 1, Name: "P1", Price: 10000}}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			Return(dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	// 11. Failed: ClearCart Error
	t.Run("failed_clear_cart_error", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, transactionMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{{ID: 1, Name: "P1", Price: 10000}}

		cartService.EXPECT().
			GetRawCart(mock.Anything, userID).
			Return(rawCart, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(products, nil)

		transactionMock.EXPECT().
			WithTransaction(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		productService.EXPECT().
			BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
				order.ID = 1
				order.InvoiceNumber = "INV-001"
				return nil
			})

		cartService.EXPECT().
			ClearCart(mock.Anything, userID).
			Return(dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestOrderUsecase_GetOrderHistory(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	dbErr := errors.New("unexpected database error")

	t.Run("success_get_history_with_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrders := []entity.Order{
			{ID: 101, UserID: userID, InvoiceNumber: "INV-001", TotalAmount: 50000, Status: "PAID"},
			{ID: 102, UserID: userID, InvoiceNumber: "INV-002", TotalAmount: 30000, Status: "PAID"},
		}
		mockItems := []entity.OrderItem{
			{OrderID: 101, ProductID: 1, ProductName: "Sepatu", Price: 25000, Quantity: 2, Subtotal: 50000},
		}

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID).
			Return(mockOrders, mockItems, nil)

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Orders, 2)

		// Order 101: Memiliki item
		assert.Equal(t, uint(101), result.Orders[0].OrderID)
		assert.Equal(t, "INV-001", result.Orders[0].InvoiceNumber)
		assert.Len(t, result.Orders[0].Items, 1)
		assert.Equal(t, "Sepatu", result.Orders[0].Items[0].ProductName)

		// Order 102: Tidak memiliki item (Memicu kondisi `if itemForThisOrder == nil`)
		assert.Equal(t, uint(102), result.Orders[1].OrderID)
		assert.Equal(t, "INV-002", result.Orders[1].InvoiceNumber)
		assert.NotNil(t, result.Orders[1].Items)
		assert.Len(t, result.Orders[1].Items, 0)
	})
	t.Run("success_get_history_empty", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID).
			Return([]entity.Order{}, []entity.OrderItem{}, nil)

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Orders, 0)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID).
			Return(nil, nil, dbErr)

		result, err := usecase.GetOrderHistory(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get order history")
		assert.ErrorContains(t, err, dbErr.Error())
	})
}

// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail"
func TestOrderUsecase_GetOrderDetail(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	orderID := uint(101)
	dbErr := errors.New("database connection failed")

	// 1. Success: Berhasil mengambil detail order beserta itemnya
	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/success_get_detail"
	t.Run("success_get_detail", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &entity.Order{
			ID:            orderID,
			UserID:        userID,
			InvoiceNumber: "INV-001",
			TotalAmount:   30000,
			Status:        "PAID",
		}
		mockItems := []entity.OrderItem{
			{OrderID: orderID, ProductID: 5, ProductName: "Kopi", Price: 15000, Quantity: 2, Subtotal: 30000},
		}

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(mockOrder, nil)

		orderRepository.EXPECT().
			FindItemsByOrderID(mock.Anything, orderID).
			Return(mockItems, nil)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderID, result.OrderID)
		assert.Equal(t, "INV-001", result.InvoiceNumber)
		assert.Equal(t, float64(30000), result.TotalAmount)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "Kopi", result.Items[0].ProductName)
	})

	// 2. Failed: Order tidak ditemukan di DB (apperror.ErrRecordNotFound)
	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_order_not_found"
	t.Run("failed_order_not_found", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(nil, apperror.ErrRecordNotFound)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})

	// 3. Failed: Unexpected error dari FindByID (Kunci Coverage 100%)
	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_unexpected_db_error_on_find_by_id"
	t.Run("failed_unexpected_db_error_on_find_by_id", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(nil, dbErr)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get order detail")
		assert.ErrorIs(t, err, dbErr)
	})

	// 4. Failed: Order milik user lain (Ownership validation fail)
	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_wrong_ownership"
	t.Run("failed_wrong_ownership", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrderWithWrongOwner := &entity.Order{
			ID:            orderID,
			UserID:        uint(99), // User ID tidak sama
			InvoiceNumber: "INV-HACKER",
		}

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(mockOrderWithWrongOwner, nil)

		// FindItemsByOrderID tidak boleh dipanggil
		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})

	// 5. Failed: Unexpected error saat mengambil items dari FindItemsByOrderID
	// go test -v ./internal/order/usecase -run "TestOrderUsecase_GetOrderDetail/failed_unexpected_db_error_on_items"
	t.Run("failed_unexpected_db_error_on_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &entity.Order{ID: orderID, UserID: userID}

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(mockOrder, nil)

		orderRepository.EXPECT().
			FindItemsByOrderID(mock.Anything, orderID).
			Return(nil, dbErr)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get order detail")
		assert.ErrorIs(t, err, dbErr)
	})
}
