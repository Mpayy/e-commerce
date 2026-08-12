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

	"github.com/Mpayy/e-commerce/database/seeder"
	"github.com/Mpayy/e-commerce/internal/product/repository"
)

// @title           E-Commerce API
// @version         1.0
// @description     Modular monolith e-commerce backend built with Go, Gin, GORM, and Redis.
// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shouldSeed := flag.Bool("seed", false, "Run database seeder if value is true")
	flag.Parse()

	application, cleanup, err := InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize API: %v", err)
	}
	defer cleanup()

	app := application.App
	router := application.Router

	if *shouldSeed {
		seeder.RunSeeder(app.Log, app.DB)
		app.Log.Info("Server is shutting down after seeding completed.")
		return
	}

	if err := repository.EnsureIndexes(ctx, app.MongoDatabase); err != nil {
		app.Log.Fatalf("Failed to ensure indexes: %v", err)
	}

	router.SetupRouter()

	host := app.Config.GetString("APP_HOST")
	port := app.Config.GetInt("APP_PORT")
	addr := fmt.Sprintf("%s:%d", host, port)

	server := &http.Server{
		Addr:    addr,
		Handler: app.Gin,
	}

	go func() {
		app.Log.Infof("Server starting on: %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	app.Log.Infof("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Log.Fatalf("Server forced to shutdown: %v", err)
	}
	app.Log.Infof("Server exited properly")
}
