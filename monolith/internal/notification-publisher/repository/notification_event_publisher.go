package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Mpayy/e-commerce/monolith/internal/notification-publisher/event"
	"github.com/Mpayy/e-commerce/monolith/internal/notification-publisher/usecase"
	"github.com/Mpayy/e-commerce/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationEventPublisher struct {
	channel *amqp.Channel
}

func NewEventPublisher(channel *amqp.Channel) usecase.EventPublisher {
	return &NotificationEventPublisher{channel: channel}
}

func (n *NotificationEventPublisher) PublishOrderCreated(ctx context.Context, event event.OrderCreatedEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order created event: %w", err)
	}

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
