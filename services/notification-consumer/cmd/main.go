package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/services/notification-consumer/dependency"
	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/delivery/consumer"
	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/repository"
	"github.com/Mpayy/e-commerce/services/notification-consumer/internal/notification/usecase"
	"golang.org/x/sync/errgroup"
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

	healthSrv := &http.Server{
		Addr: ":8084",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"UP"}`))
		}),
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return orderConsumer.Start(gCtx)
	})

	g.Go(func() error {
		log.Infof("health check server running on %s", healthSrv.Addr)
		if err := healthSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return healthSrv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		log.Errorf("notification-consumer stopped with error: %v", err)
	}
	log.Info("notification-consumer shut down gracefully")
}
