package dependency

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewSqlxDB(pool *pgxpool.Pool) *sqlx.DB {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return sqlx.NewDb(sqlDB, "pgx")
}
