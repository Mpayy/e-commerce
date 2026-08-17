package dependency

import (
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewGormDB(pool *pgxpool.Pool, cfg *config.Config, log *logger.Logger) (*gorm.DB, error) {
	sqlDB := stdlib.OpenDBFromPool(pool)

	gormLogLevel := gormlogger.Warn
	if cfg.AppEnv != "production" {
		gormLogLevel = gormlogger.Info
	}

	customGormLogger := gormlogger.New(
		&logrusWriter{Log: log},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  gormLogLevel,
		},
	)

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:         customGormLogger,
		TranslateError: true,
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

type logrusWriter struct {
	Log *logger.Logger
}

func (l *logrusWriter) Printf(message string, args ...any) {
	l.Log.Tracef(message, args...)
}
