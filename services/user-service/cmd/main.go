package main

import (
	"context"
	"errors"
	"flag"
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
	seeder "github.com/Mpayy/e-commerce/services/user-service/database/seeder"
	"github.com/Mpayy/e-commerce/services/user-service/dependency"
	userhttp "github.com/Mpayy/e-commerce/services/user-service/internal/user/delivery/http"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/repository"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/usecase"
	"github.com/gin-gonic/gin"
)

func setupRouter(r *gin.Engine, userHttp userhttp.UserHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/register", userHttp.Register)
		api.POST("/login", userHttp.Login)

		auth := api.Group("", authMiddleware.RequireAuth())
		auth.GET("/profile", userHttp.GetProfile)
		auth.DELETE("/logout", userHttp.Logout)
	}
	return r
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	shouldSeed := flag.Bool("seed", false, "Run database seeder if value is true")
	flag.Parse()

	cfg := config.Load()
	log := logger.NewLogger(cfg)
	jwtToken := jwt.NewJwtToken(cfg)
	validator := validator.NewValidator()
	engine := engine.NewGin(cfg, log)
	pool, dbCleanup, err := dependency.NewPostgresPool(cfg, log)
	if err != nil {
		log.Fatalf("failed to initialize postgres pool: %v", err)
	}
	defer dbCleanup()

	sqlxDB := dependency.NewSqlxDB(pool)
	if *shouldSeed {
		seeder.RunSeeder(log, sqlxDB)
		log.Info("Server is shutting down after seeding completed.")
		return
	}

	rdb, rdbCleanup := cache.NewRedisCli(cfg, log)
	defer rdbCleanup()

	redisRepo := repository.NewUserRedisRepository(rdb)
	repo := repository.NewUserRepository(sqlxDB)
	userUsecase := usecase.NewUserUsecase(repo, redisRepo, log, jwtToken)
	userHttp := userhttp.NewUserHandler(userUsecase, validator)

	sessionChecker := middleware.NewRedisSessionChecker(rdb)
	authMiddleware := middleware.NewAuthMiddleware(jwtToken, sessionChecker, log)
	router := setupRouter(engine, userHttp, authMiddleware)
	srv := &http.Server{
		Addr:    ":8082",
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
