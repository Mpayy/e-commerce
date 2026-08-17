package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityLogRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewActivityLogRepository(pool *pgxpool.Pool) ActivityLogRepository {
	return &ActivityLogRepositoryImpl{pool: pool}
}

func (r *ActivityLogRepositoryImpl) Create(ctx context.Context, log *entity.ActivityLog) error {
	query := `INSERT INTO activity_logs (order_id, user_id, message, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (order_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, log.OrderID, log.UserID, log.Message, log.CreatedAt)
	return err
}
