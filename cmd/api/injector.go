//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Mpayy/e-commerce/dependency"
	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/internal/user/delivery/http"
	"github.com/Mpayy/e-commerce/internal/user/repository"
	"github.com/Mpayy/e-commerce/internal/user/usecase"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/Mpayy/e-commerce/routes"
	"github.com/google/wire"
)

var userSet = wire.NewSet(
	repository.NewUserRepository,
	usecase.NewUserUsecase,
	http.NewUserHandler,
)

var middlewareSet = wire.NewSet(
	middleware.NewAuthMiddleware,
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
