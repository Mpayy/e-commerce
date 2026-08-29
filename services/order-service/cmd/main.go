package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/pkg/cache"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/engine"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/validator"
	"github.com/Mpayy/e-commerce/services/order-service/dependency"
	_ "github.com/Mpayy/e-commerce/services/order-service/docs"
	cartHttp "github.com/Mpayy/e-commerce/services/order-service/internal/cart/delivery/http"
	cartRepo "github.com/Mpayy/e-commerce/services/order-service/internal/cart/repository"
	cartUC "github.com/Mpayy/e-commerce/services/order-service/internal/cart/usecase"
	orderHttp "github.com/Mpayy/e-commerce/services/order-service/internal/order/delivery/http"
	orderRepo "github.com/Mpayy/e-commerce/services/order-service/internal/order/repository"
	orderUC "github.com/Mpayy/e-commerce/services/order-service/internal/order/usecase"
	productRepo "github.com/Mpayy/e-commerce/services/order-service/internal/product/repository"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter(r *gin.Engine, orderHandler orderHttp.OrderHandler, cartHandler cartHttp.CartHandler, AuthMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api/v1", AuthMiddleware.RequireAuth())
	{
		cart := api.Group("/cart")
		cart.POST("", cartHandler.AddItem)
		cart.GET("", cartHandler.GetCart)
		cart.PATCH("/:product_id", cartHandler.UpdateItem)
		cart.DELETE("/:product_id", cartHandler.RemoveItem)
		cart.DELETE("", cartHandler.ClearCart)

		order := api.Group("/orders")
		order.POST("", orderHandler.Checkout)
		order.GET("", orderHandler.GetHistory)
		order.GET("/:order_id", orderHandler.GetDetail)

		admin := api.Group("/admin", AuthMiddleware.RequireAuth(), middleware.AdminMiddleware())
		admin.GET("/analytics/sales", orderHandler.GetSalesAnalytics)
	}

	return r
}

// @title           Order & Cart Service API
// @version         1.0
// @description     Shopping cart (Redis Hash) and checkout/order history (PostgreSQL via sqlc). Resolves product data via gRPC to the Product Service; publishes an order.created event to RabbitMQ after successful checkout.
// @host            localhost:8083
// @BasePath        /api/v1

// @securitydefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.NewLogger(cfg)
	jwtToken := jwt.NewJwtToken(cfg)
	validator := validator.NewValidator()
	engine := engine.NewGin(cfg, log)
	pool, cleanupDB, err := dependency.NewPostgresPool(cfg, log)
	if err != nil {
		log.Fatalf("failed to initialize postgres pool: %v", err)
	}
	defer cleanupDB()

	rdb, cleanupRedis := cache.NewRedisCli(cfg, log)
	defer cleanupRedis()

	grpcConn, cleanupGRPC, err := dependency.NewProductServiceConn(cfg, log)
	if err != nil {
		log.Fatalf("failed to initialize product service gRPC connection: %v", err)
	}
	defer cleanupGRPC()

	channel, cleanupMQ, err := dependency.NewRabbitMQChannel(cfg, log)
	if err != nil {
		log.Fatalf("failed to initialize rabbitmq channel: %v", err)
	}
	defer cleanupMQ()

	eventPublisher := orderRepo.NewNotificationEventPublisher(channel)
	productClient := productRepo.NewProductGRPCClient(grpcConn)
	orderRepository := orderRepo.NewOrderRepository(pool)
	cartRepository := cartRepo.NewCartRedisRepository(rdb)

	cartUsecase := cartUC.NewCartUsecase(cartRepository, productClient, log)
	orderUsecase := orderUC.NewOrderUsecase(orderRepository, log, cartUsecase, productClient, eventPublisher)

	cartHandler := cartHttp.NewCartHandler(cartUsecase, cartUsecase, validator)
	orderHandler := orderHttp.NewOrderHandler(orderUsecase, validator)

	sessionChecker := middleware.NewRedisSessionChecker(rdb)
	authMiddleware := middleware.NewAuthMiddleware(jwtToken, sessionChecker, log)
	router := setupRouter(engine, orderHandler, cartHandler, authMiddleware)
	srv := &http.Server{
		Addr:    ":8083",
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
}
