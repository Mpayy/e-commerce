package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Mpayy/e-commerce/services/order-service/internal/order/event"
	"github.com/Mpayy/e-commerce/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationEventPublisher struct {
	channel *amqp.Channel
	mu      sync.Mutex
}

func NewNotificationEventPublisher(channel *amqp.Channel) *NotificationEventPublisher {
	return &NotificationEventPublisher{channel: channel}
}

func (n *NotificationEventPublisher) PublishOrderCreated(ctx context.Context, event event.OrderCreatedEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order created event: %w", err)
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	err = n.channel.PublishWithContext(ctx, messaging.OrderEventsExchange, messaging.OrderCreatedRoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         msg,
	})

	if err != nil {
		return fmt.Errorf("failed to publish order created event: %w", err)
	}

	return nil
}
