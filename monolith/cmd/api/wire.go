//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Mpayy/e-commerce/monolith/dependency"
	carthttp "github.com/Mpayy/e-commerce/monolith/internal/cart/delivery/http"
	cartrepository "github.com/Mpayy/e-commerce/monolith/internal/cart/repository"
	cartusecase "github.com/Mpayy/e-commerce/monolith/internal/cart/usecase"
	publisherrepository "github.com/Mpayy/e-commerce/monolith/internal/notification-publisher/repository"
	orderhttp "github.com/Mpayy/e-commerce/monolith/internal/order/delivery/http"
	orderrepository "github.com/Mpayy/e-commerce/monolith/internal/order/repository"
	orderusecase "github.com/Mpayy/e-commerce/monolith/internal/order/usecase"
	productrepository "github.com/Mpayy/e-commerce/monolith/internal/product/repository"
	"github.com/Mpayy/e-commerce/monolith/internal/routes"
	userhttp "github.com/Mpayy/e-commerce/monolith/internal/user/delivery/http"
	userrepository "github.com/Mpayy/e-commerce/monolith/internal/user/repository"
	userusecase "github.com/Mpayy/e-commerce/monolith/internal/user/usecase"
	"github.com/Mpayy/e-commerce/pkg/cache"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/database"
	"github.com/Mpayy/e-commerce/pkg/engine"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/messaging"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/Mpayy/e-commerce/pkg/validator"
	productv1 "github.com/Mpayy/e-commerce/proto/product/v1"
	"github.com/google/wire"
)

var userSet = wire.NewSet(
	userrepository.NewUserRedisRepository,
	wire.Bind(new(middleware.SessionChecker), new(userrepository.UserRedisRepository)),
	userrepository.NewUserRepository,
	userusecase.NewUserUsecase,
	userhttp.NewUserHandler,
)

var productSet = wire.NewSet(
	dependency.NewProductServiceConn,
	productv1.NewProductServiceClient,
	productrepository.NewProductGRPCClient,
)

var cartSet = wire.NewSet(
	cartrepository.NewCartRedisRepository,
	cartusecase.NewCartUsecase,
	wire.Bind(new(cartusecase.CartService), new(*cartusecase.CartUsecaseImpl)),
	wire.Bind(new(cartusecase.CartUsecase), new(*cartusecase.CartUsecaseImpl)),
	carthttp.NewCartHandler,
)

var orderSet = wire.NewSet(
	orderrepository.NewOrderRepository,
	orderusecase.NewOrderUsecase,
	orderhttp.NewOrderHandler,
)

var publisherSet = wire.NewSet(
	publisherrepository.NewEventPublisher,
)

var InfrastructureSet = wire.NewSet(
	config.Load,
	logger.NewLogger,
	validator.NewValidator,
	cache.NewRedisCli,
	engine.NewGin,
	database.NewPostgresDB,
	messaging.NewRabbitMQConn,
	dependency.NewGormDB,
	dependency.NewRabbitMQChannel,
)

func InitializeApp() (*dependency.App, func(), error) {
	wire.Build(
		InfrastructureSet,
		userSet,
		productSet,
		cartSet,
		orderSet,
		publisherSet,
		middleware.NewAuthMiddleware,
		routes.NewRouter,
		jwt.NewJwtToken,
		transaction.NewTransaction,
		dependency.NewApp,
	)
	return nil, nil, nil
}
