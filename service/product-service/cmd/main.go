package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/pkg/cache"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/database"
	"github.com/Mpayy/e-commerce/pkg/engine"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	productv1 "github.com/Mpayy/e-commerce/proto/product/v1"
	productgrpc "github.com/Mpayy/e-commerce/service/product-service/internal/product/delivery/grpc"
	producthttp "github.com/Mpayy/e-commerce/service/product-service/internal/product/delivery/http"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/repository"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
)

func setupRouter(r *gin.Engine, categoryHandler producthttp.CategoryHandler, productHandler producthttp.ProductHandler, AuthMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/categories", categoryHandler.GetAll)
		api.GET("/products", productHandler.Search)
		api.GET("/products/:product_id", productHandler.GetByID)

		admin := api.Group("/admin", AuthMiddleware.RequireAuth(), middleware.AdminMiddleware())
		admin.POST("/categories", categoryHandler.Create)
		admin.POST("/products", productHandler.Create)
		admin.PATCH("/products/:product_id", productHandler.Update)
		admin.DELETE("/products/:product_id", productHandler.Delete)
		admin.PATCH("/products/:product_id/adjust-stock", productHandler.AdjustStock)
	}

	return r
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()
	log := logger.NewLogger(cfg)
	jwtToken := jwt.NewJwtToken(cfg)
	engine := engine.NewGin(cfg, log)
	rdb, rdbCleanup := cache.NewRedisCli(cfg, log)
	defer rdbCleanup()
	db, cleanup, err := database.NewMongoDB(cfg, log)
	if err != nil {
		log.Fatalf("failed to initialize mongodb: %v", err)
	}
	defer cleanup()

	if err := repository.EnsureIndexes(ctx, db); err != nil {
		log.Errorf("failed to ensure indexes: %v", err)
		return
	}

	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	sessionRepo := repository.NewSessionRepository(rdb)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, log)
	productUsecase := usecase.NewProductUsecase(productRepo, categoryUsecase, log)
	validator := validator.New()
	categoryHandler := producthttp.NewCategoryHandler(categoryUsecase, validator)
	productHandler := producthttp.NewProductHandler(productUsecase, validator)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	productGrpcServer := productgrpc.NewProductGRPCServer(productUsecase)
	productv1.RegisterProductServiceServer(grpcServer, productGrpcServer)

	go func() {
		log.Info("gRPC server running on :50051")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	authMiddleware := middleware.NewAuthMiddleware(jwtToken, sessionRepo, log)
	router := setupRouter(engine, categoryHandler, productHandler, authMiddleware)
	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		log.Infof("HTTP Server running on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down HTTP server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")

	grpcServer.GracefulStop()
	log.Info("gRPC server stopped")
}
