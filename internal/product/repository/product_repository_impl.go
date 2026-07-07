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
	"gorm.io/gorm/clause"
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

func (r *ProductRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	var product entity.Product
	err := r.GetTx(ctx).First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepositoryImpl) FindByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	var products []entity.Product
	err := r.GetTx(ctx).Where("id IN ?", ids).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepositoryImpl) FindAll(ctx context.Context, filter *entity.ProductFilter) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64
	query := r.GetTx(ctx).Model(&entity.Product{}).Where("is_active = ?", true)
	if filter.Search != "" {
		query = query.Where("name LIKE ?", "%"+filter.Search+"%")
	}
	if filter.CategoryID != 0 {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.Limit
	err := query.Limit(filter.Limit).Offset(offset).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entity.Product) error {
	result := r.GetTx(ctx).Model(product).Select("category_id", "name", "slug", "description", "price", "stock", "sku", "is_active").Updates(product)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			msg := strings.ToLower(mysqlErr.Message)
			if strings.Contains(msg, "products.slug") {
				return apperror.ErrDuplicatedProduct
			}

			if strings.Contains(msg, "products.sku") {
				return apperror.ErrDuplicatedProductSku
			}
		}
		return result.Error
	}
	return nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.GetTx(ctx).Model(&entity.Product{}).Where("id = ? AND is_active = ?", id, true).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *ProductRepositoryImpl) DecreaseStock(ctx context.Context, productID uint, quantity int) error {
	var product entity.Product
	if err := r.GetTx(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}
	if product.Stock < quantity {
		return apperror.ErrInsufficientStock
	}
	r.GetTx(ctx).Model(&product).Update("stock", product.Stock-quantity)
	return nil
}

func (r *ProductRepositoryImpl) AdjustStock(ctx context.Context, productID uint, quantity int) error {
	result := r.GetTx(ctx).Model(&entity.Product{}).Where("id = ?", productID).Update("stock", quantity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
