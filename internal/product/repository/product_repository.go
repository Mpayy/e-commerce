package productrepository

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
}