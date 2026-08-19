package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/services/product-service/internal/product/dto"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_category_usecase.go
type CategoryUsecase interface {
	CreateCategory(ctx context.Context, request *dto.CategoryRequest) (*dto.CategoryResponse, error)
	GetAllCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	ValidateCategoryExists(ctx context.Context, id uint) error
}
