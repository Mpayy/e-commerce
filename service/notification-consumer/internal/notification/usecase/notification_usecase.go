package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/dto"
)

type NotificationUsecase interface {
	HandleOrderCreated(ctx context.Context, event dto.OrderCreatedEvent) error
}
