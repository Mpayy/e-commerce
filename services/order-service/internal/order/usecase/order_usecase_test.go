package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"io"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	cartMock "github.com/Mpayy/e-commerce/services/order-service/internal/cart/mocks"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/entity"
	repoMock "github.com/Mpayy/e-commerce/services/order-service/internal/order/mocks"
	productentity "github.com/Mpayy/e-commerce/services/order-service/internal/product/entity"
	productMock "github.com/Mpayy/e-commerce/services/order-service/internal/product/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupOrderUsecase(t *testing.T) (OrderUsecase, *productMock.MockProductService, *cartMock.MockCartService, *repoMock.MockOrderRepository, *repoMock.MockEventPublisher) {
	orderRepository := repoMock.NewMockOrderRepository(t)
	cartService := cartMock.NewMockCartService(t)
	productService := productMock.NewMockProductService(t)
	publisherMock := repoMock.NewMockEventPublisher(t)
	log := newTestLogger()

	t.Cleanup(func() {
		orderRepository.AssertExpectations(t)
		cartService.AssertExpectations(t)
		productService.AssertExpectations(t)
		publisherMock.AssertExpectations(t)
	})

	orderUsecase := NewOrderUsecase(orderRepository, log, cartService, productService, publisherMock)
	return orderUsecase, productService, cartService, orderRepository, publisherMock
}

func TestOrderUsecase_Checkout(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	dbErr := errors.New("unexpected database error")
	grpcErr := errors.New("grpc connection error")

	t.Run("success_checkout", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, publisherMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2, 2: 1}
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
			{ID: 2, Name: "Produk 2", Price: 20000, Stock: 5},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
				order.ID = 100
				order.InvoiceNumber = "INV-20260824-000100"
				return nil
			})

		cartService.EXPECT().ClearCart(mock.Anything, userID).Return(nil)
		publisherMock.EXPECT().PublishOrderCreated(mock.Anything, mock.Anything).Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(100), result.OrderID)
		assert.Equal(t, "INV-20260824-000100", result.InvoiceNumber)
		assert.Equal(t, "PAID", result.Status)
		assert.Equal(t, float64(40000), result.TotalAmount)
		assert.Len(t, result.Items, 2)
	})

	t.Run("success_checkout_with_non_fatal_errors", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, publisherMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 1}
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
				order.ID = 101
				order.InvoiceNumber = "INV-101"
				return nil
			})

		cartService.EXPECT().ClearCart(mock.Anything, userID).Return(errors.New("redis timeout"))
		publisherMock.EXPECT().PublishOrderCreated(mock.Anything, mock.Anything).Return(errors.New("rabbitmq down"))

		result, err := usecase.Checkout(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(101), result.OrderID)
	})

	t.Run("success_checkout_skipping_zero_qty_item", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, publisherMock := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2, 2: 0} // Product 2 qty = 0 (diproses skip)
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
			{ID: 2, Name: "Produk 2", Price: 20000, Stock: 5},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		orderRepository.EXPECT().
			CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
				order.ID = 102
				return nil
			})

		cartService.EXPECT().ClearCart(mock.Anything, userID).Return(nil)
		publisherMock.EXPECT().PublishOrderCreated(mock.Anything, mock.Anything).Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 1)
	})

	t.Run("failed_get_raw_cart", func(t *testing.T) {
		usecase, _, cartService, _, _ := setupOrderUsecase(t)

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(nil, dbErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("failed_cart_empty_initial", func(t *testing.T) {
		usecase, _, cartService, _, _ := setupOrderUsecase(t)

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(map[uint]int{}, nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCartEmpty)
	})

	t.Run("failed_get_products_by_ids", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(nil, grpcErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, grpcErr)
	})

	t.Run("failed_product_not_found_in_map", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2, 99: 1} // ID 99 tidak di-return oleh grpc
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_cart_empty_after_zero_qty_filtering", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 0} // Semua item ber-qty 0
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 5},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCartEmpty)
	})

	t.Run("failed_bulk_decrease_stock", func(t *testing.T) {
		usecase, productService, cartService, _, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(grpcErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, grpcErr)
	})

	t.Run("failed_create_order_restore_stock_success", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		orderRepository.EXPECT().CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).Return(dbErr)
		productService.EXPECT().BulkRestoreStock(mock.Anything, mock.Anything).Return(nil)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to create order with items")
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("failed_create_order_restore_stock_failed", func(t *testing.T) {
		usecase, productService, cartService, orderRepository, _ := setupOrderUsecase(t)

		rawCart := map[uint]int{1: 2}
		products := []productentity.Product{
			{ID: 1, Name: "Produk 1", Price: 10000, Stock: 10},
		}

		cartService.EXPECT().GetRawCart(mock.Anything, userID).Return(rawCart, nil)
		productService.EXPECT().GetProductsByIDs(mock.Anything, mock.Anything).Return(products, nil)
		productService.EXPECT().BulkDecreaseStock(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		orderRepository.EXPECT().CreateOrderWithItems(mock.Anything, mock.Anything, mock.Anything).Return(dbErr)
		restoreErr := errors.New("failed to reach grpc for restore")
		productService.EXPECT().BulkRestoreStock(mock.Anything, mock.Anything).Return(restoreErr)

		result, err := usecase.Checkout(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to create order with items")
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestOrderUsecase_GetOrderHistory(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	dbErr := errors.New("unexpected database error")

	t.Run("success_default_pagination_and_mapping_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		filter := &dto.OrderFilter{
			Page:  0,
			Limit: 0,
		}

		mockOrders := []entity.Order{
			{
				ID:            101,
				UserID:        userID,
				InvoiceNumber: "INV-001",
				TotalAmount:   50000,
				Status:        "PAID",
				Items: []entity.OrderItem{
					{ProductID: 1, ProductName: "Sepatu", Price: 25000, Quantity: 2, Subtotal: 50000},
				},
			},
		}
		totalCount := int64(1)

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID, 1, 10).
			Return(mockOrders, totalCount, nil)

		result, err := usecase.GetOrderHistory(ctx, userID, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, filter.Page)
		assert.Equal(t, 10, filter.Limit)
		assert.Len(t, result.Orders, 1)

		assert.Equal(t, 1, result.Meta.Page)
		assert.Equal(t, 10, result.Meta.Limit)
		assert.Equal(t, int64(1), result.Meta.Total)
		assert.Equal(t, int64(1), result.Meta.TotalPages)

		orderRes := result.Orders[0]
		assert.Equal(t, uint(101), orderRes.OrderID)
		assert.Equal(t, "INV-001", orderRes.InvoiceNumber)
		assert.Equal(t, 50000.0, orderRes.TotalAmount)
		assert.Equal(t, "PAID", orderRes.Status)
		assert.Len(t, orderRes.Items, 1)

		itemRes := orderRes.Items[0]
		assert.Equal(t, uint(1), itemRes.ProductID)
		assert.Equal(t, "Sepatu", itemRes.ProductName)
		assert.Equal(t, 25000.0, itemRes.Price)
		assert.Equal(t, 2, itemRes.Quantity)
		assert.Equal(t, 50000.0, itemRes.Subtotal)
	})

	t.Run("success_custom_pagination_and_empty_items", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		filter := &dto.OrderFilter{
			Page:  2,
			Limit: 5,
		}

		mockOrders := []entity.Order{
			{
				ID:            102,
				UserID:        userID,
				InvoiceNumber: "INV-002",
				TotalAmount:   30000,
				Status:        "PENDING",
				Items:         []entity.OrderItem{},
			},
		}
		totalCount := int64(12)

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID, 2, 5).
			Return(mockOrders, totalCount, nil)

		result, err := usecase.GetOrderHistory(ctx, userID, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Meta.Page)
		assert.Equal(t, 5, result.Meta.Limit)
		assert.Equal(t, int64(12), result.Meta.Total)
		assert.Equal(t, int64(3), result.Meta.TotalPages)

		assert.Len(t, result.Orders, 1)
		assert.Len(t, result.Orders[0].Items, 0)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		filter := &dto.OrderFilter{
			Page:  1,
			Limit: 10,
		}

		orderRepository.EXPECT().
			FindByUserID(mock.Anything, userID, 1, 10).
			Return(nil, int64(0), dbErr)

		result, err := usecase.GetOrderHistory(ctx, userID, filter)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get order history")
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestOrderUsecase_GetOrderDetail(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	orderID := uint(101)
	dbErr := errors.New("database connection failed")

	t.Run("success_get_detail", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &entity.Order{
			ID:            orderID,
			UserID:        userID,
			InvoiceNumber: "INV-001",
			TotalAmount:   30000,
			Status:        "PAID",
			Items: []entity.OrderItem{
				{OrderID: orderID, ProductID: 5, ProductName: "Kopi", Price: 15000, Quantity: 2, Subtotal: 30000},
			},
		}

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(mockOrder, nil)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderID, result.OrderID)
		assert.Equal(t, "INV-001", result.InvoiceNumber)
		assert.Equal(t, float64(30000), result.TotalAmount)
		assert.Equal(t, "PAID", result.Status)
		assert.Len(t, result.Items, 1)

		// Verifikasi mapping items
		assert.Equal(t, uint(5), result.Items[0].ProductID)
		assert.Equal(t, "Kopi", result.Items[0].ProductName)
		assert.Equal(t, 15000.0, result.Items[0].Price)
		assert.Equal(t, 2, result.Items[0].Quantity)
		assert.Equal(t, 30000.0, result.Items[0].Subtotal)
	})

	t.Run("failed_order_not_found", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(nil, apperror.ErrRecordNotFound)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})

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

	t.Run("failed_wrong_ownership", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrderWithWrongOwner := &entity.Order{
			ID:            orderID,
			UserID:        uint(99),
			InvoiceNumber: "INV-HACKER",
			Items:         []entity.OrderItem{},
		}

		orderRepository.EXPECT().
			FindByID(mock.Anything, orderID).
			Return(mockOrderWithWrongOwner, nil)

		result, err := usecase.GetOrderDetail(ctx, userID, orderID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})
}

func TestOrderUsecase_GetSalesAnalytics(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("database connection failed")

	// Sample mock data untuk skenario sukses
	mockDailyRevenue := []entity.DailyRevenueRow{
		{
			Date:         time.Date(2026, 8, 20, 0, 0, 0, 0, time.Local),
			OrderCount:   5,
			DailyRevenue: 150000,
			RunningTotal: 150000,
		},
		{
			Date:         time.Date(2026, 8, 21, 0, 0, 0, 0, time.Local),
			OrderCount:   3,
			DailyRevenue: 90000,
			RunningTotal: 240000,
		},
	}

	mockTopProducts := []entity.TopProductRow{
		{
			Rank:              1,
			ProductID:         10,
			ProductName:       "Kopi Susu",
			TotalQuantitySold: 20,
			TotalRevenue:      200000,
		},
		{
			Rank:              2,
			ProductID:         11,
			ProductName:       "Roti Bakar",
			TotalQuantitySold: 8,
			TotalRevenue:      40000,
		},
	}

	t.Run("success_with_explicit_dates_and_limit", func(t *testing.T) {
		usecase, _, _, orderRepo, _ := setupOrderUsecase(t)

		req := &dto.SalesAnalyticsRequest{
			From:  "2026-08-01",
			To:    "2026-08-25",
			Limit: 10,
		}

		// Mock repository calls
		orderRepo.EXPECT().
			GetDailyRevenueReport(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(mockDailyRevenue, nil).
			Once()

		orderRepo.EXPECT().
			GetTopProducts(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), int32(10)).
			Return(mockTopProducts, nil).
			Once()

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Assert Period
		assert.Equal(t, "2026-08-01", res.Period.From)
		assert.Equal(t, "2026-08-25", res.Period.To)

		// Assert Summary Accumulation (150000 + 90000 = 240000 | 5 + 3 = 8)
		assert.Equal(t, float64(240000), res.Summary.TotalRevenue)
		assert.Equal(t, int64(8), res.Summary.TotalOrders)

		// Assert Daily Revenue Mapping
		assert.Len(t, res.DailyRevenue, 2)
		assert.Equal(t, "2026-08-20", res.DailyRevenue[0].Date)
		assert.Equal(t, int64(5), res.DailyRevenue[0].OrderCount)

		// Assert Top Products Mapping
		assert.Len(t, res.TopProducts, 2)
		assert.Equal(t, uint(10), res.TopProducts[0].ProductID)
		assert.Equal(t, "Kopi Susu", res.TopProducts[0].ProductName)
	})

	t.Run("success_with_empty_dates_and_default_limit", func(t *testing.T) {
		usecase, _, _, orderRepo, _ := setupOrderUsecase(t)

		// Empty request -> trigger default limit=5 & 30 days date fallback
		req := &dto.SalesAnalyticsRequest{
			From:  "",
			To:    "",
			Limit: 0,
		}

		orderRepo.EXPECT().
			GetDailyRevenueReport(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]entity.DailyRevenueRow{}, nil).
			Once()

		orderRepo.EXPECT().
			GetTopProducts(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), int32(5)). // Limit defaulted to 5
			Return([]entity.TopProductRow{}, nil).
			Once()

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, float64(0), res.Summary.TotalRevenue)
		assert.Equal(t, int64(0), res.Summary.TotalOrders)
		assert.Empty(t, res.DailyRevenue)
		assert.Empty(t, res.TopProducts)
	})

	t.Run("error_invalid_to_date_format", func(t *testing.T) {
		usecase, _, _, _, _ := setupOrderUsecase(t)

		req := &dto.SalesAnalyticsRequest{
			From: "2026-08-01",
			To:   "25-08-2026", // Wrong format
		}

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrInvalidDateRange)
	})

	t.Run("error_invalid_from_date_format", func(t *testing.T) {
		usecase, _, _, _, _ := setupOrderUsecase(t)

		req := &dto.SalesAnalyticsRequest{
			From: "invalid-date", // Wrong format
			To:   "2026-08-25",
		}

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrInvalidDateRange)
	})

	t.Run("error_get_daily_revenue_report_repository_failed", func(t *testing.T) {
		usecase, _, _, orderRepo, _ := setupOrderUsecase(t)

		req := &dto.SalesAnalyticsRequest{
			From: "2026-08-01",
			To:   "2026-08-25",
		}

		// Simulasikan error pada salah satu goroutine
		orderRepo.EXPECT().
			GetDailyRevenueReport(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(nil, dbErr).
			Once()

		orderRepo.EXPECT().
			GetTopProducts(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("int32")).
			Return(mockTopProducts, nil).
			Maybe() // Gunakan Maybe() karena errgroup bisa menyelesaikan atau membatalkan context goroutine ini

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to get daily revenue report")
	})

	t.Run("error_get_top_products_repository_failed", func(t *testing.T) {
		usecase, _, _, orderRepo, _ := setupOrderUsecase(t)

		req := &dto.SalesAnalyticsRequest{
			From: "2026-08-01",
			To:   "2026-08-25",
		}

		orderRepo.EXPECT().
			GetDailyRevenueReport(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(mockDailyRevenue, nil).
			Maybe()

		orderRepo.EXPECT().
			GetTopProducts(mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("int32")).
			Return(nil, dbErr).
			Once()

		res, err := usecase.GetSalesAnalytics(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to get top products")
	})
}

func TestOrderUsecase_GetAdminOrderList(t *testing.T) {
	ctx := context.Background()
	dummyErr := errors.New("database connection failed")

	t.Run("success_with_default_pagination_and_empty_dates", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			Page:  0,
			Limit: 0,
		}

		now := time.Now()
		mockOrders := []entity.Order{
			{
				ID:            1,
				UserID:        10,
				InvoiceNumber: "INV-001",
				TotalAmount:   150000,
				Status:        "PAID",
				CreatedAt:     now,
			},
		}

		orderRepository.EXPECT().
			GetAdminOrderList(mock.Anything, mock.MatchedBy(func(filter *entity.OrderFilter) bool {
				return filter.Page == 1 && filter.Limit == 10 && filter.From == nil && filter.To == nil
			})).
			Return(mockOrders, nil)

		orderRepository.EXPECT().
			CountAdminOrderList(mock.Anything, mock.MatchedBy(func(filter *entity.OrderFilter) bool {
				return filter.Page == 1 && filter.Limit == 10
			})).
			Return(int64(1), nil)

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 1, res.Meta.Page)
		assert.Equal(t, 10, res.Meta.Limit)
		assert.Equal(t, int64(1), res.Meta.Total)
		assert.Equal(t, int64(1), res.Meta.TotalPages)
		assert.Len(t, res.Orders, 1)
		assert.Equal(t, uint(1), res.Orders[0].OrderID)
		assert.Equal(t, "INV-001", res.Orders[0].InvoiceNumber)
		assert.Equal(t, now.Format(time.RFC3339), res.Orders[0].CreatedAt)
	})

	t.Run("success_with_valid_dates_and_zero_total_records", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			From:  "2026-09-01",
			To:    "2026-09-10",
			Page:  2,
			Limit: 5,
		}

		orderRepository.EXPECT().
			GetAdminOrderList(mock.Anything, mock.MatchedBy(func(filter *entity.OrderFilter) bool {
				if filter.From == nil || filter.To == nil {
					return false
				}
				expectedFrom := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
				expectedTo := time.Date(2026, 9, 10, 23, 59, 59, 999999999, time.UTC)
				return filter.From.Equal(expectedFrom) && filter.To.Equal(expectedTo) && filter.Page == 2 && filter.Limit == 5
			})).
			Return([]entity.Order{}, nil)

		orderRepository.EXPECT().
			CountAdminOrderList(mock.Anything, mock.Anything).
			Return(int64(0), nil)

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(0), res.Meta.Total)
		assert.Equal(t, int64(0), res.Meta.TotalPages)
		assert.Empty(t, res.Orders)
	})

	t.Run("error_invalid_to_date_format", func(t *testing.T) {
		usecase, _, _, _, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			To: "invalid-date",
		}

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrInvalidDateRange)
	})

	t.Run("error_invalid_from_date_format", func(t *testing.T) {
		usecase, _, _, _, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			From: "01-09-2026",
		}

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrInvalidDateRange)
	})

	t.Run("error_get_admin_order_list_repo_fails", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			Page:  1,
			Limit: 10,
		}

		orderRepository.EXPECT().
			GetAdminOrderList(mock.Anything, mock.Anything).
			Return(nil, dummyErr)

		orderRepository.EXPECT().
			CountAdminOrderList(mock.Anything, mock.Anything).
			Return(int64(0), nil).
			Maybe()

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get admin order list")
	})

	t.Run("error_count_admin_order_list_repo_fails", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		req := &dto.AdminOrderListRequest{
			Page:  1,
			Limit: 10,
		}

		orderRepository.EXPECT().
			GetAdminOrderList(mock.Anything, mock.Anything).
			Return([]entity.Order{}, nil).
			Maybe()

		orderRepository.EXPECT().
			CountAdminOrderList(mock.Anything, mock.Anything).
			Return(int64(0), dummyErr)

		res, err := usecase.GetAdminOrderList(ctx, req)

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count admin order list")
	})
}

func TestOrderUsecase_GetAdminOrderDetail(t *testing.T) {
	ctx := context.Background()
	orderID := uint(145)
	dbErr := errors.New("unexpected database error")

	t.Run("success_get_admin_order_detail", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		mockOrder := &entity.Order{
			ID:            145,
			UserID:        10,
			InvoiceNumber: "INV-20260825-000145",
			TotalAmount:   750000,
			Status:        "PAID",
			Items: []entity.OrderItem{
				{
					ID:          1,
					OrderID:     145,
					ProductID:   12,
					ProductName: "Mechanical Keyboard 60%",
					Price:       750000,
					Quantity:    1,
					Subtotal:    750000,
				},
			},
		}

		orderRepository.EXPECT().FindByID(ctx, orderID).Return(mockOrder, nil)

		result, err := usecase.GetAdminOrderDetail(ctx, orderID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(145), result.OrderID)
		assert.Equal(t, "INV-20260825-000145", result.InvoiceNumber)
		assert.Equal(t, float64(750000), result.TotalAmount)
		assert.Equal(t, "PAID", result.Status)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, uint(12), result.Items[0].ProductID)
		assert.Equal(t, "Mechanical Keyboard 60%", result.Items[0].ProductName)
		assert.Equal(t, float64(750000), result.Items[0].Price)
		assert.Equal(t, 1, result.Items[0].Quantity)
		assert.Equal(t, float64(750000), result.Items[0].Subtotal)
	})

	t.Run("error_order_not_found", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().FindByID(ctx, orderID).Return(nil, apperror.ErrRecordNotFound)

		result, err := usecase.GetAdminOrderDetail(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrOrderNotFound)
	})

	t.Run("error_repository_unexpected_error", func(t *testing.T) {
		usecase, _, _, orderRepository, _ := setupOrderUsecase(t)

		orderRepository.EXPECT().FindByID(ctx, orderID).Return(nil, dbErr)

		result, err := usecase.GetAdminOrderDetail(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed find by id")
		assert.ErrorIs(t, err, dbErr)
	})
}