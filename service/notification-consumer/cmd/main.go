package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/service/notification-consumer/dependency"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/delivery/consumer"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/repository"
	"github.com/Mpayy/e-commerce/service/notification-consumer/internal/notification/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.NewLogger(cfg)

	pool, cleanupDB, err := dependency.NewPostgresPool(cfg, log)
	if err != nil {
		log.Fatalf("failed initialize postgres: %v", err)
	}
	defer cleanupDB()

	channel, cleanupMQ, err := dependency.NewRabbitMQChannel(cfg, log)
	if err != nil {
		log.Fatalf("failed initialize rabbitmq: %v", err)
	}
	defer cleanupMQ()

	repo := repository.NewActivityLogRepository(pool)
	notifUsecase := usecase.NewNotificationUsecase(repo, log)
	orderConsumer := consumer.NewOrderCreatedConsumer(channel, notifUsecase, log)

	if err := orderConsumer.Start(ctx); err != nil {
		log.Errorf("consumer stopped with error: %v", err)
	}

	log.Info("notification-consumer shutting down")
}
