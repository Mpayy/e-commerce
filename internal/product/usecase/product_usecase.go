package productusecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/dto"
)

type ProductUsecase interface {
	CreateProduct(ctx context.Context, request *dto.ProductCreateRequest) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, id uint, request *dto.ProductUpdateRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uint) error
	SearchProducts(ctx context.Context, request *dto.ProductSearchRequest) (*dto.ProductSearchResponse, error)
	GetProductDetail(ctx context.Context, id uint) (*dto.ProductResponse, error)
	AdjustStock(ctx context.Context, productID uint, quantity int) error
}