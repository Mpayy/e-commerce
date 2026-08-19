package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/dto"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/repository"
	"github.com/gosimple/slug"
)

type CategoryUsecaseImpl struct {
	CategoryRepo repository.CategoryRepository
	Log          *logger.Logger
}

func NewCategoryUsecase(categoryRepo repository.CategoryRepository, log *logger.Logger) CategoryUsecase {
	return &CategoryUsecaseImpl{CategoryRepo: categoryRepo, Log: log}
}

func (u *CategoryUsecaseImpl) CreateCategory(ctx context.Context, request *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	logger := u.Log.WithField("name", request.Name)
	logger.Debug("Attempting to create category")

	category := &entity.Category{
		Name: request.Name,
		Slug: slug.Make(request.Name),
	}

	if err := u.CategoryRepo.Create(ctx, category); err != nil {
		if errors.Is(err, apperror.ErrDuplicatedKey) {
			logger.WithField("name", request.Name).Warn("Create category failed: duplicate name")
			return nil, apperror.ErrDuplicatedCategory
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	response := &dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}

	u.Log.WithField("name", request.Name).Info("Category created successfully")
	return response, nil
}

func (u *CategoryUsecaseImpl) GetAllCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	u.Log.Debug("Attempting to get all categories")

	categories, err := u.CategoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all categories: %w", err)
	}
	responses := []dto.CategoryResponse{}
	for _, category := range categories {
		responses = append(responses, dto.CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		})
	}

	u.Log.WithField("count", len(responses)).Info("Found all categories successfully")
	return responses, nil
}

func (u *CategoryUsecaseImpl) ValidateCategoryExists(ctx context.Context, id uint) error {
	logger := u.Log.WithField("id", id)
	logger.Debug("Attempting to validate category existence")

	_, err := u.CategoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			logger.Warn("Failed to validate category: category not found")
			return apperror.ErrCategoryNotFound
		}
		return fmt.Errorf("failed to validate category: %w", err)
	}

	logger.Debug("Category validated successfully")
	return nil
}
