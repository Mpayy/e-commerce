package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(name string, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, func(), error) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=%s",
			cfg.DatabaseHost,
			cfg.DatabaseUsername,
			cfg.DatabasePassword,
			cfg.DatabasePort,
			cfg.DatabaseSSLMode,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

	poolConfig.ConnConfig.Database = name
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to ping pgxpool: %w", err)
	}

	cleanup := func() {
		pool.Close()
		log.Info("pgxpool connection closed successfully")
	}

	return pool, cleanup, nil
}
