package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/entity"
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
	DecreaseStock(ctx context.Context, productID uint, quantity int) error
	AdjustStock(ctx context.Context, productID uint, quantity int) error
}
