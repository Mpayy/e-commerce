package engine

import (
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func NewGin(cfg *config.Config, log *logger.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.LoggerMiddleware(log))

	return engine
}