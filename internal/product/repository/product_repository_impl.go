package productrepository

import (
	"context"
	"errors"
	"strings"

	"github.com/Mpayy/e-commerce/internal/product/entity"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ProductRepositoryImpl struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &ProductRepositoryImpl{DB: db}
}

func (r *ProductRepositoryImpl) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.GetTxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return r.DB.WithContext(ctx)
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product *entity.Product) error {
	err := r.GetTx(ctx).Create(product).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			msg := strings.ToLower(mysqlErr.Message)
			if strings.Contains(msg, "products.slug") {
				return apperror.ErrDuplicatedProduct
			}
			if strings.Contains(msg, "products.sku") {
				return apperror.ErrDuplicatedProductSku
			}
		}
		return err
	}
	return nil
}
