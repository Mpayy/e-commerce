package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_product_repository.go
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id uint) (*entity.Product, error)
	FindByIDs(ctx context.Context, ids []uint) ([]entity.Product, error)
	FindAll(ctx context.Context, filter *entity.ProductFilter) ([]entity.Product, int64, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uint) error
	BulkDecreaseStock(ctx context.Context, idemKey string, items []entity.StockItem) error
	BulkRestoreStock(ctx context.Context, idemKey string, items []entity.StockItem) error
	AdjustStock(ctx context.Context, productID uint, quantity int) error
	UpdateImagePath(ctx context.Context, productID uint, newImagePath string) error
}
