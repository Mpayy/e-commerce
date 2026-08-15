package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
)

func NewPostgresDB(cfg *config.Config, log *logger.Logger) (*sql.DB, func(), error) {
	host := cfg.DatabaseHost
	user := cfg.DatabaseUsername
	password := cfg.DatabasePassword
	dbname := cfg.DatabaseName
	port := cfg.DatabasePort
	sslmode := cfg.DatabaseSSLMode

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		host, user, password, dbname, port, sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			log.Errorf("failed to close postgres connection: %v", err)
		}
	}

	return db, cleanup, nil
}
