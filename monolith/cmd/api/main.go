package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/monolith/database/seeder"
)

// @title           E-Commerce API
// @version         1.0
// @description     Modular monolith e-commerce backend built with Go, Gin, GORM, and Redis.
// @host            localhost:8080
// @BasePath        /api/v1
// @securitydefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shouldSeed := flag.Bool("seed", false, "Run database seeder if value is true")
	flag.Parse()

	app, cleanup, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize API: %v", err)
	}
	defer cleanup()

	cfg := app.Cfg
	logger := app.Log

	if *shouldSeed {
		seeder.RunSeeder(logger, app.DB)
		logger.Info("Server is shutting down after seeding completed.")
		return
	}

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
