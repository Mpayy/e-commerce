package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/messaging"
	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/dto"
	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedConsumerImpl struct {
	channel *amqp.Channel
	usecase usecase.NotificationUsecase
	log     *logger.Logger
}

func NewOrderCreatedConsumer(channel *amqp.Channel, usecase usecase.NotificationUsecase, log *logger.Logger) OrderCreatedConsumer {
	return &OrderCreatedConsumerImpl{channel: channel, usecase: usecase, log: log}
}

func (c *OrderCreatedConsumerImpl) Start(ctx context.Context) error {
	if err := c.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}

	msgs, err := c.channel.Consume(
		messaging.NotificationQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer stopping: context cancelled")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				c.log.Info("consumer stopping: channel closed")
				return nil
			}

			var event dto.OrderCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				c.log.WithError(err).Error("failed to parse event, message discarded")
				msg.Nack(false, false)
				continue
			}

			if err := c.usecase.HandleOrderCreated(ctx, event); err != nil {
				c.log.WithError(err).Error("event processing failed, it will be requeued")
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}
}
