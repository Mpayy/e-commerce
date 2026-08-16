package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/monolith/internal/product/entity"
)

type ProductService interface {
	GetByProductID(ctx context.Context, id uint) (*entity.Product, error)
	GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error)
	BulkDecreaseStock(ctx context.Context, checkoutID string, items []entity.BulkDecreaseStock) error
	BulkRestoreStock(ctx context.Context, checkoutID string) error
}
