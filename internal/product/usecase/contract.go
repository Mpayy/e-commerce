package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/entity"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_product_service.go
type ProductService interface {
	GetByProductID(ctx context.Context, id uint) (*entity.Product, error)
	GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error)
	DecreaseStock(ctx context.Context, productID uint, quantity int) error
}
