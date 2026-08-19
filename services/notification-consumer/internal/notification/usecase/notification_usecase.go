package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/dto"
)

type NotificationUsecase interface {
	HandleOrderCreated(ctx context.Context, event dto.OrderCreatedEvent) error
}
