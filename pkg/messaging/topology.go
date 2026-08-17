package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	OrderEventsExchange    = "order.events"
	OrderCreatedRoutingKey = "order.created"
	NotificationQueue      = "notification.order.created"
)

func DeclareOrderEventsTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(OrderEventsExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", OrderEventsExchange, err)
	}

	q, err := ch.QueueDeclare(NotificationQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", NotificationQueue, err)
	}

	if err := ch.QueueBind(q.Name, OrderCreatedRoutingKey, OrderEventsExchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", q.Name, OrderEventsExchange, err)
	}

	return nil
}
