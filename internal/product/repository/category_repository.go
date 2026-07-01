package productrepository

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindAll(ctx context.Context) ([]*entity.Category, error)
}
