package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title           E-Commerce Monolith API
// @version         1.0
// @description     User authentication, shopping cart, and checkout service — part of a microservices e-commerce system. Backed by PostgreSQL and Redis. Product and category data is resolved via gRPC calls to a separate Product Service (see :8081/swagger/index.html); successful checkouts publish an order.created event to RabbitMQ for async notification handling.
// @host            localhost:8080
// @BasePath        /api/v1
// @securitydefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, cleanup, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize API: %v", err)
	}
	defer cleanup()

	cfg := app.Cfg
	logger := app.Log

	engine := app.Router.SetupRouter()

	host := cfg.AppHost
	port := cfg.AppPort
	addr := fmt.Sprintf("%s:%s", host, port)

	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		logger.Infof("Server starting on: %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	logger.Infof("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Infof("Server exited properly")
}
