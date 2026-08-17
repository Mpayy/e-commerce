package dependency

import (
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQChannel(cfg *config.Config, log *logger.Logger) (*amqp.Channel, func(), error) {
	conn, cleanupConn, err := messaging.NewRabbitMQConn(cfg, log)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		cleanupConn()
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	if err := messaging.DeclareOrderEventsTopology(ch); err != nil {
		cleanupConn()
		return nil, nil, fmt.Errorf("failed to declare topology: %w", err)
	}

	cleanup := func() {
		if err := ch.Close(); err != nil {
			log.Errorf("failed to close rabbitmq channel: %v", err)
		}
		log.Infof("RabbitMQ channel closed successfully")
		cleanupConn()
	}

	return ch, cleanup, nil
}