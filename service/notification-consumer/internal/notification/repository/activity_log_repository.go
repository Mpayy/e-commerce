package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/entity"
)

type ActivityLogRepository interface {
	Create(ctx context.Context, log *entity.ActivityLog) error
}