package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, func(), error) {
	host := cfg.DatabaseHost
	user := cfg.DatabaseUsername
	password := cfg.DatabasePassword
	dbname := cfg.DatabaseName
	port := cfg.DatabasePort
	sslmode := cfg.DatabaseSSLMode

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		host, user, password, dbname, port, sslmode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

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
