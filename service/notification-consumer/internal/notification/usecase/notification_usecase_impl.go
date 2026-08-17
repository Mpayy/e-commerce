package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/dto"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/entity"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/repository"
)

type NotificationUsecaseImpl struct {
	repository repository.ActivityLogRepository
	log        *logger.Logger
}

func NewNotificationUsecase(repo repository.ActivityLogRepository, log *logger.Logger) NotificationUsecase {
	return &NotificationUsecaseImpl{repository: repo, log: log}
}

func (u *NotificationUsecaseImpl) HandleOrderCreated(ctx context.Context, event dto.OrderCreatedEvent) error {
	u.log.Infof("[SIMULATED EMAIL] To user #%d: Pesanan %s berhasil dibuat, total Rp%.0f",
		event.UserID, event.InvoiceNumber, event.TotalAmount)

	activity := &entity.ActivityLog{
		OrderID:   event.OrderID,
		UserID:    event.UserID,
		Message:   fmt.Sprintf("Order %s created", event.InvoiceNumber),
		CreatedAt: time.Now(),
	}

	return u.repository.Create(ctx, activity)
}
