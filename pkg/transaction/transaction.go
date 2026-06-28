package transaction

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "tx"

type Transaction interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionImpl struct {
	DB *gorm.DB
}

func NewTransaction(db *gorm.DB) Transaction {
	return &TransactionImpl{DB: db}
}

func (t *TransactionImpl) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.DB.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

func GetTxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	return tx, ok
}
