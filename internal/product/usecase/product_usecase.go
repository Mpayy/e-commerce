package productusecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/dto"
)

type ProductUsecase interface {
	CreateProduct(ctx context.Context, request *dto.ProductRequest) (*dto.ProductResponse, error)
}