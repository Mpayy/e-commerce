package dependency

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type App struct {
	Gin *gin.Engine
	Log *logrus.Logger
	Config *viper.Viper
	DB *gorm.DB
}

func NewApp(gin *gin.Engine, log *logrus.Logger, config *viper.Viper, db *gorm.DB) *App {
	return &App{
		Gin: gin,
		Log: log,
		Config: config,
		DB: db,
	}
}