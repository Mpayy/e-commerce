package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
	cartUC "github.com/Mpayy/e-commerce/services/order-service/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/entity"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/event"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/repository"
	productentity "github.com/Mpayy/e-commerce/services/order-service/internal/product/entity"
	productUC "github.com/Mpayy/e-commerce/services/order-service/internal/product/usecase"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type OrderUsecaseImpl struct {
	orderRepository repository.OrderRepository
	log             *logger.Logger
	cartService     cartUC.CartService
	productService  productUC.ProductService
	eventPublisher  EventPublisher
}

func NewOrderUsecase(orderRepository repository.OrderRepository, log *logger.Logger, cartService cartUC.CartService, productService productUC.ProductService, eventPublisher EventPublisher) OrderUsecase {
	return &OrderUsecaseImpl{
		orderRepository: orderRepository,
		log:             log,
		cartService:     cartService,
		productService:  productService,
		eventPublisher:  eventPublisher,
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

	productIDs := make([]uint, 0, len(rawCart))
	for productID := range rawCart {
		productIDs = append(productIDs, productID)
	}

	products, err := u.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[uint]productentity.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}

	checkoutID := uuid.NewString()
	orderItems := make([]entity.OrderItem, 0, len(rawCart))
	decrements := make([]productentity.BulkDecreaseStock, 0, len(rawCart))
	var grandTotal float64

	for productID, qty := range rawCart {
		if qty <= 0 {
			continue
		}

		product, exists := productMap[productID]
		if !exists {
			return nil, apperror.ErrProductNotFound
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
		return nil, apperror.ErrCartEmpty
	}

	if err = u.productService.BulkDecreaseStock(ctx, checkoutID, decrements); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to create order with items: %w", err)
	}

	if err = u.cartService.ClearCart(ctx, userID); err != nil {
		log.WithError(err).Error("failed to clear cart after successful checkout")
	}

	orderEvent := event.OrderCreatedEvent{
		OrderID:       order.ID,
		UserID:        order.UserID,
		InvoiceNumber: order.InvoiceNumber,
		TotalAmount:   order.TotalAmount,
	}

	if err = u.eventPublisher.PublishOrderCreated(ctx, orderEvent); err != nil {
		log.WithError(err).Error("failed to publish order.created event, but checkout remain successful")
	}

	responseItems := make([]dto.OrderItemResponse, 0, len(orderItems))
	for _, item := range orderItems {
		responseItems = append(responseItems, dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	log.WithFields(logger.Fields{
		"order_id":     order.ID,
		"total_amount": order.TotalAmount,
		"items":        len(responseItems),
	}).Info("Checkout successful")

	return &dto.OrderResponse{
		OrderID:       order.ID,
		InvoiceNumber: order.InvoiceNumber,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		Items:         responseItems,
	}, nil
}

func (u *OrderUsecaseImpl) GetOrderHistory(ctx context.Context, userID uint, filter *dto.OrderFilter) (*dto.OrderHistoryResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Getting order history")

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	orders, total, err := u.orderRepository.FindByUserID(ctx, userID, filter.Page, filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get order history: %w", err)
	}

	var totalPages int64
	if total > 0 {
		totalPages = (total + int64(filter.Limit) - 1) / int64(filter.Limit)
	}

	responseOrders := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {

		itemsDTO := make([]dto.OrderItemResponse, 0, len(order.Items))
		for _, item := range order.Items {
			itemsDTO = append(itemsDTO, dto.OrderItemResponse{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
				Subtotal:    item.Subtotal,
			})
		}

		responseOrders = append(responseOrders, dto.OrderResponse{
			OrderID:       order.ID,
			InvoiceNumber: order.InvoiceNumber,
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			Items:         itemsDTO,
		})
	}

	log.WithFields(logger.Fields{
		"user_id":     userID,
		"order_count": len(responseOrders),
	}).Debug("Order history retrieved successfully")

	return &dto.OrderHistoryResponse{
		Orders: responseOrders,
		Meta: dto.MetaPagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
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

	responseItems := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
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

func (u *OrderUsecaseImpl) GetSalesAnalytics(ctx context.Context, req *dto.SalesAnalyticsRequest) (*dto.SalesAnalyticsResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"from":  req.From,
		"to":    req.To,
		"limit": req.Limit,
	})
	log.Debug("Getting Sales Analytics")

	now := time.Now()
	var toTime time.Time
	if req.To == "" {
		toTime = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)
	} else {
		parsed, err := time.Parse("2006-01-02", req.To)
		if err != nil {
			return nil, apperror.ErrInvalidDateRange
		}
		toTime = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, time.UTC)
	}

	var fromTime time.Time
	if req.From == "" {
		thirtyDaysAgo := toTime.AddDate(0, 0, -30)
		fromTime = time.Date(thirtyDaysAgo.Year(), thirtyDaysAgo.Month(), thirtyDaysAgo.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		parsed, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			return nil, apperror.ErrInvalidDateRange
		}
		fromTime = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}

	var (
		dailyRevenueReport []entity.DailyRevenueRow
		topProducts        []entity.TopProductRow
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		dailyRevenueReport, err = u.orderRepository.GetDailyRevenueReport(gCtx, fromTime, toTime)
		if err != nil {
			return fmt.Errorf("failed to get daily revenue report: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		topProducts, err = u.orderRepository.GetTopProducts(gCtx, fromTime, toTime, int32(limit))
		if err != nil {
			return fmt.Errorf("failed to get top products: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var totalRevenue float64
	var totalOrders int64
	dailyRevenueRes := make([]dto.DailyRevenueResponse, 0, len(dailyRevenueReport))

	for _, revenue := range dailyRevenueReport {
		totalRevenue += revenue.DailyRevenue
		totalOrders += revenue.OrderCount

		dailyRevenueRes = append(dailyRevenueRes, dto.DailyRevenueResponse{
			Date:         revenue.Date.Format("2006-01-02"),
			OrderCount:   revenue.OrderCount,
			DailyRevenue: revenue.DailyRevenue,
			RunningTotal: revenue.RunningTotal,
		})
	}

	summaryRes := dto.SummaryResponse{
		TotalRevenue: totalRevenue,
		TotalOrders:  totalOrders,
	}

	topProductRes := make([]dto.TopProductResponse, 0, len(topProducts))
	for _, product := range topProducts {
		topProductRes = append(topProductRes, dto.TopProductResponse{
			Rank:              product.Rank,
			ProductID:         product.ProductID,
			ProductName:       product.ProductName,
			TotalQuantitySold: product.TotalQuantitySold,
			TotalRevenue:      product.TotalRevenue,
		})
	}

	log.Debug("Sales Analytics retrieved successfully")
	return &dto.SalesAnalyticsResponse{
		Period: dto.PeriodResponse{
			From: fromTime.Format("2006-01-02"),
			To:   toTime.Format("2006-01-02"),
		},
		Summary:      summaryRes,
		DailyRevenue: dailyRevenueRes,
		TopProducts:  topProductRes,
	}, nil
}

func (u *OrderUsecaseImpl) GetAdminOrderList(ctx context.Context, req *dto.AdminOrderListRequest) (*dto.AdminOrderListResponse, error) {
	var toTime *time.Time
	if req.To != "" {
		parsed, err := time.Parse("2006-01-02", req.To)
		if err != nil {
			return nil, apperror.ErrInvalidDateRange
		}
		t := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, time.UTC)
		toTime = &t
	}

	var fromTime *time.Time
	if req.From != "" {
		parsed, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			return nil, apperror.ErrInvalidDateRange
		}
		t := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		fromTime = &t
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	filter := &entity.OrderFilter{
		UserID:    req.UserID,
		Status:    req.Status,
		MinAmount: req.MinAmount,
		MaxAmount: req.MaxAmount,
		From:      fromTime,
		To:        toTime,
		Page:      page,
		Limit:     limit,
	}

	var (
		orders       []entity.Order
		totalRecords int64
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		orders, err = u.orderRepository.GetAdminOrderList(gCtx, filter)
		if err != nil {
			return fmt.Errorf("failed to get admin order list: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		totalRecords, err = u.orderRepository.CountAdminOrderList(gCtx, filter)
		if err != nil {
			return fmt.Errorf("failed to count admin order list: %w", err)
		}
		return nil

	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var totalPages int64
	if totalRecords > 0 {
		totalPages = (totalRecords + int64(limit) - 1) / int64(limit)
	}

	orderSummaryRes := make([]dto.AdminOrderSummaryResponse, 0, len(orders))
	for _, order := range orders {
		orderSummaryRes = append(orderSummaryRes, dto.AdminOrderSummaryResponse{
			OrderID:       order.ID,
			UserID:        order.UserID,
			InvoiceNumber: order.InvoiceNumber,
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.AdminOrderListResponse{
		Orders: orderSummaryRes,
		Meta: dto.MetaPagination{
			Page:       page,
			Limit:      limit,
			Total:      totalRecords,
			TotalPages: totalPages,
		},
	}, nil
}

func (u *OrderUsecaseImpl) GetAdminOrderDetail(ctx context.Context, orderID uint) (*dto.OrderResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"order_id": orderID,
	})
	log.Debug("Getting admin order detail")

	order, err := u.orderRepository.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed find by id: %w", err)
	}

	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	log.Debug("Admin order detail retrieved successfully")

	return &dto.OrderResponse{
		OrderID:       order.ID,
		InvoiceNumber: order.InvoiceNumber,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		Items:         items,
	}, nil
}
