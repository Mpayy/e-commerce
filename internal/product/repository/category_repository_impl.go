package productrepository

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/internal/product/entity"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type CategoryRepositoryImpl struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &CategoryRepositoryImpl{DB: db}
}

func (r *CategoryRepositoryImpl) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.GetTxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return r.DB.WithContext(ctx)
}

func (r *CategoryRepositoryImpl) Create(ctx context.Context, category *entity.Category) error {
	err := r.GetTx(ctx).Create(category).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return apperror.ErrDuplicatedKey
		}
		return err
	}

	return nil
}

func (r *CategoryRepositoryImpl) FindAll(ctx context.Context) ([]*entity.Category, error) {
	var categories []*entity.Category
	err := r.GetTx(ctx).Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.Category, error) {
	var category entity.Category
	err := r.GetTx(ctx).First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &category, nil
}
