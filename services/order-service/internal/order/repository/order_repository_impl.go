package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/entity"
	sqlcgen "github.com/Mpayy/e-commerce/services/order-service/internal/order/repository/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepositoryImpl struct {
	pool    *pgxpool.Pool
	queries *sqlcgen.Queries
}

func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &OrderRepositoryImpl{pool: pool, queries: sqlcgen.New(pool)}
}

func (r *OrderRepositoryImpl) CreateOrderWithItems(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.queries.WithTx(tx)

	created, err := qtx.CreateOrder(ctx, sqlcgen.CreateOrderParams{
		UserID:      int64(order.UserID),
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
	})
	if err != nil {
		return err
	}

	invoiceNumber := fmt.Sprintf("INV-%s-%06d", created.CreatedAt.Format("20060102"), created.ID)
	if err := qtx.UpdateInvoiceNumber(ctx, sqlcgen.UpdateInvoiceNumberParams{
		ID: created.ID,
		InvoiceNumber: pgtype.Text{
			String: invoiceNumber,
			Valid:  true,
		},
	}); err != nil {
		return err
	}

	for _, item := range items {
		if _, err := qtx.CreateOrderItem(ctx, sqlcgen.CreateOrderItemParams{
			OrderID:     created.ID,
			ProductID:   int64(item.ProductID),
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    int32(item.Quantity),
			Subtotal:    item.Subtotal,
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	order.ID = uint(created.ID)
	order.InvoiceNumber = invoiceNumber
	return nil
}

func (r *OrderRepositoryImpl) FindByUserID(ctx context.Context, userID uint, page int, limit int) ([]entity.Order, int64, error) {
	offset := (page - 1) * limit

	sqlcOrders, err := r.queries.ListOrdersByUser(ctx, sqlcgen.ListOrdersByUserParams{
		UserID: int64(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountOrdersByUser(ctx, int64(userID))
	if err != nil {
		return nil, 0, err
	}

	if len(sqlcOrders) == 0 {
		return []entity.Order{}, total, nil
	}

	var orderIDs []int64
	for _, order := range sqlcOrders {
		orderIDs = append(orderIDs, order.ID)
	}

	sqlcItems, err := r.queries.ListItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, 0, err
	}

	itemsByOrderID := make(map[int64][]entity.OrderItem)
	for _, item := range sqlcItems {
		itemsByOrderID[item.OrderID] = append(itemsByOrderID[item.OrderID], entity.OrderItem{
			ID:          uint(item.ID),
			OrderID:     uint(item.OrderID),
			ProductID:   uint(item.ProductID),
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    int(item.Quantity),
			Subtotal:    item.Subtotal,
		})
	}

	orders := make([]entity.Order, 0, len(sqlcOrders))
	for _, order := range sqlcOrders {
		items := itemsByOrderID[order.ID]
		if items == nil {
			items = []entity.OrderItem{}
		}

		orders = append(orders, entity.Order{
			ID:            uint(order.ID),
			InvoiceNumber: order.InvoiceNumber.String,
			UserID:        uint(order.UserID),
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt,
			UpdatedAt:     order.UpdatedAt,
			Items:         items,
		})
	}

	return orders, total, nil
}

func (r *OrderRepositoryImpl) FindByID(ctx context.Context, orderID uint) (*entity.Order, error) {
	sqlcOrder, err := r.queries.GetOrderByID(ctx, int64(orderID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}

	sqlcItems, err := r.queries.GetItemsByOrderID(ctx, int64(orderID))
	if err != nil {
		return nil, err
	}

	items := make([]entity.OrderItem, 0, len(sqlcItems))
	for _, item := range sqlcItems {
		items = append(items, entity.OrderItem{
			ID:          uint(item.ID),
			OrderID:     uint(item.OrderID),
			ProductID:   uint(item.ProductID),
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    int(item.Quantity),
			Subtotal:    item.Subtotal,
		})
	}

	order := entity.Order{
		ID:            uint(sqlcOrder.ID),
		UserID:        uint(sqlcOrder.UserID),
		InvoiceNumber: sqlcOrder.InvoiceNumber.String,
		TotalAmount:   sqlcOrder.TotalAmount,
		Status:        sqlcOrder.Status,
		CreatedAt:     sqlcOrder.CreatedAt,
		UpdatedAt:     sqlcOrder.UpdatedAt,
		Items:         items,
	}

	return &order, nil
}

func (r *OrderRepositoryImpl) GetDailyRevenueReport(ctx context.Context, from, to time.Time) ([]entity.DailyRevenueRow, error) {
	sqlcGetDailyRevenueReportRow, err := r.queries.GetDailyRevenueReport(ctx, sqlcgen.GetDailyRevenueReportParams{
		FromDate: from,
		ToDate:   to,
	})

	if err != nil {
		return nil, err
	}

	if len(sqlcGetDailyRevenueReportRow) == 0 {
		return []entity.DailyRevenueRow{}, nil
	}

	getDailyRevenueReportRow := make([]entity.DailyRevenueRow, 0, len(sqlcGetDailyRevenueReportRow))
	for _, row := range sqlcGetDailyRevenueReportRow {
		getDailyRevenueReportRow = append(getDailyRevenueReportRow, entity.DailyRevenueRow{
			Date:         row.Date,
			OrderCount:   row.OrderCount,
			DailyRevenue: row.DailyRevenue,
			RunningTotal: row.RunningTotal,
		})
	}

	return getDailyRevenueReportRow, nil
}

func (r *OrderRepositoryImpl) GetTopProducts(ctx context.Context, from, to time.Time, limit int32) ([]entity.TopProductRow, error) {
	sqlcGetTopProductsRow, err := r.queries.GetTopProducts(ctx, sqlcgen.GetTopProductsParams{
		LimitCount: limit,
		FromDate:   from,
		ToDate:     to,
	})

	if err != nil {
		return nil, err
	}

	if len(sqlcGetTopProductsRow) == 0 {
		return []entity.TopProductRow{}, nil
	}

	getTopProductsRow := make([]entity.TopProductRow, 0, len(sqlcGetTopProductsRow))
	for _, row := range sqlcGetTopProductsRow {
		getTopProductsRow = append(getTopProductsRow, entity.TopProductRow{
			ProductID:         uint(row.ProductID),
			ProductName:       row.ProductName,
			TotalQuantitySold: row.TotalQuantitySold,
			TotalRevenue:      row.TotalRevenue,
			Rank:              row.Rank,
		})
	}

	return getTopProductsRow, nil
}

func (r *OrderRepositoryImpl) applyOrderFilter(builder sq.SelectBuilder, filter *entity.OrderFilter) sq.SelectBuilder {
	if filter.Status != "" {
		builder = builder.Where(sq.Eq{"status": filter.Status})
	}
	if filter.MinAmount != 0 {
		builder = builder.Where(sq.GtOrEq{"total_amount": filter.MinAmount})
	}
	if filter.MaxAmount != 0 {
		builder = builder.Where(sq.LtOrEq{"total_amount": filter.MaxAmount})
	}
	if filter.UserID != 0 {
		builder = builder.Where(sq.Eq{"user_id": filter.UserID})
	}
	if filter.From != nil {
		builder = builder.Where(sq.GtOrEq{"created_at": filter.From})
	}
	if filter.To != nil {
		builder = builder.Where(sq.LtOrEq{"created_at": filter.To})
	}
	return builder
}

func (r *OrderRepositoryImpl) GetAdminOrderList(ctx context.Context, filter *entity.OrderFilter) ([]entity.Order, error) {
	builder := sq.Select("id, user_id, invoice_number, total_amount, status, created_at").From("orders").PlaceholderFormat(sq.Dollar)

	builder = r.applyOrderFilter(builder, filter)

	offset := (filter.Page - 1) * filter.Limit

	query, args, err := builder.OrderBy("created_at DESC").Limit(uint64(filter.Limit)).Offset(uint64(offset)).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]entity.Order, 0)
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.InvoiceNumber,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
		); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) CountAdminOrderList(ctx context.Context, filter *entity.OrderFilter) (int64, error) {
	builder := sq.Select("COUNT(*)").
		From("orders").
		PlaceholderFormat(sq.Dollar)

	builder = r.applyOrderFilter(builder, filter)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var totalRecords int64
	err = r.pool.QueryRow(ctx, query, args...).Scan(&totalRecords)
	if err != nil {
		return 0, err
	}

	return totalRecords, nil
}

func (r *OrderRepositoryImpl) CancelOrder(ctx context.Context, orderID uint) (*entity.Order, error) {
	sqlcCancelOrderRow, err := r.queries.CancelOrder(ctx, int64(orderID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, fetchErr := r.queries.GetOrderByID(ctx, int64(orderID))
			if fetchErr != nil {
				if errors.Is(fetchErr, sql.ErrNoRows) {
					return nil, apperror.ErrRecordNotFound
				}
				return nil, fetchErr
			}
			return nil, apperror.ErrStatusTransitionFailed
		}
		return nil, err
	}

	order := &entity.Order{
		ID:            uint(sqlcCancelOrderRow.ID),
		InvoiceNumber: sqlcCancelOrderRow.InvoiceNumber.String,
		Status:        sqlcCancelOrderRow.Status,
	}

	return order, nil
}
