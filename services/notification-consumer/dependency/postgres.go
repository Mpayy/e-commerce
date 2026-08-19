package dependency

import (
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/database"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, func(), error) {
	pool, cleanup, err := database.NewPostgresDB(cfg, log)
	if err != nil {
		return nil, nil, err
	}

	return pool, cleanup, nil
}
