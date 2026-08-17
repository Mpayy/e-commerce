package dependency

import (
	"github.com/Mpayy/e-commerce/monolith/internal/routes"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"gorm.io/gorm"
	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	Router       *routes.Router
	DB           *gorm.DB
	Cfg          *config.Config
	Log          *logger.Logger
	Channel      *amqp.Channel
}

func NewApp(router *routes.Router, db *gorm.DB, cfg *config.Config, log *logger.Logger, channel *amqp.Channel) *App {
	return &App{
		Router: router,
		DB:     db,
		Cfg:    cfg,
		Log:    log,
		Channel: channel,
	}
}
