package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/monolith/internal/notification-publisher/event"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_event_publisher.go
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event event.OrderCreatedEvent) error
}
