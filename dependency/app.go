package dependency

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type App struct {
	Gin    *gin.Engine
	Log    *logrus.Logger
	Config *viper.Viper
	DB     *gorm.DB
	Redis  *redis.Client
}

func NewApp(gin *gin.Engine, log *logrus.Logger, config *viper.Viper, db *gorm.DB, redis *redis.Client) *App {
	return &App{
		Gin:    gin,
		Log:    log,
		Config: config,
		DB:     db,
		Redis:  redis,
	}
}
