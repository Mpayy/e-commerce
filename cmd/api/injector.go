//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Mpayy/e-commerce/dependency"
	"github.com/Mpayy/e-commerce/internal/middleware"
	producthttp "github.com/Mpayy/e-commerce/internal/product/delivery/http"
	productrepository "github.com/Mpayy/e-commerce/internal/product/repository"
	productusecase "github.com/Mpayy/e-commerce/internal/product/usecase"
	userhttp "github.com/Mpayy/e-commerce/internal/user/delivery/http"
	userrepository "github.com/Mpayy/e-commerce/internal/user/repository"
	userusecase "github.com/Mpayy/e-commerce/internal/user/usecase"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/Mpayy/e-commerce/routes"
	"github.com/google/wire"
)

var userSet = wire.NewSet(
	userrepository.NewUserRepository,
	userusecase.NewUserUsecase,
	userhttp.NewUserHandler,
)

var categorySet = wire.NewSet(
	productrepository.NewCategoryRepository,
	productusecase.NewCategoryUsecase,
	producthttp.NewCategoryHandler,
)

var productSet = wire.NewSet(
	productrepository.NewProductRepository,
	productusecase.NewProductUsecase,
	producthttp.NewProductHandler,
)

var middlewareSet = wire.NewSet(
	middleware.NewAuthMiddleware,
	middleware.NewAdminMiddleware,
)

var routeSet = wire.NewSet(
	routes.NewRouter,
)

func InitializeApplication() *Application {
	wire.Build(
		// Dependency
		dependency.NewViper,
		dependency.NewGorm,
		dependency.NewRedis,
		dependency.NewValidator,
		dependency.NewLogrus,
		dependency.NewGin,

		// User
		userSet,

		// Category
		categorySet,

		// Product
		productSet,

		// Middleware
		middlewareSet,

		// Route
		routeSet,

		// JWT
		jwt.NewJwtToken,

		// Transaction
		transaction.NewTransaction,

		// App
		dependency.NewApp,

		// Injector
		NewApplication,
	)
	return nil
}
