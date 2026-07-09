package usecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/product/dto"
)

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, request *dto.CategoryRequest) (*dto.CategoryResponse, error)
	GetAllCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	ValidateCategoryExists(ctx context.Context, id uint) error
}
