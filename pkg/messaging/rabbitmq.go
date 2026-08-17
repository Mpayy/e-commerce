package messaging

import (
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQConn(cfg *config.Config, log *logger.Logger) (*amqp.Connection, func(), error) {
	url := cfg.RabbitMQUrl

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed connect to rabbitmq: %w", err)
	}

	cleanup := func() {
		if err := conn.Close(); err != nil {
			log.Errorf("failed to close rabbitmq connection: %v", err)
		}
		log.Infof("RabbitMQ connection closed successfully")
	}

	return conn, cleanup, nil
}
