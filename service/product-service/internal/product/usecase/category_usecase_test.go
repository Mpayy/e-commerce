package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/dto"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/service/product-service/internal/product/mocks"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLoggerCategory() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupCategoryUsecase(t *testing.T) (CategoryUsecase, *mocks.MockCategoryRepository) {
	categoryRepository := mocks.NewMockCategoryRepository(t)
	log := newTestLoggerCategory()
	usecase := NewCategoryUsecase(categoryRepository, log)
	t.Cleanup(func() {
		categoryRepository.AssertExpectations(t)
	})
	return usecase, categoryRepository
}

// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory"
func TestCategoryUsecaseImpl_CreateCategory(t *testing.T) {
	ctx := context.Background()
	request := &dto.CategoryRequest{
		Name: "Test Category",
	}
	requestSlug := slug.Make(request.Name)
	dbErr := errors.New("unexpected error")

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/successful_create_category"
	t.Run("successful_create_category", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			return u != nil && u.Name == request.Name && u.Slug == requestSlug
		})).Return(nil)

		res, err := uc.CreateCategory(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, request.Name, res.Name)
		assert.Equal(t, requestSlug, res.Slug)
	})

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/failed_to_create_category_duplicate"
	t.Run("failed_to_create_category_duplicate", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			return u != nil && u.Name == request.Name && u.Slug == requestSlug
		})).Return(apperror.ErrDuplicatedKey)

		res, err := uc.CreateCategory(ctx, request)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedCategory)
	})

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			return u != nil && u.Name == request.Name && u.Slug == requestSlug
		})).Return(dbErr)

		res, err := uc.CreateCategory(ctx, request)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, dbErr)
	})
}

// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_GetAllCategories"
func TestCategoryUsecaseImpl_GetAllCategories(t *testing.T) {
	ctx := context.Background()
	categories := []entity.Category{
		{
			ID:   uint(1),
			Name: "Test Category 1",
			Slug: "test-category-1",
		},
		{
			ID:   uint(2),
			Name: "Test Category 2",
			Slug: "test-category-2",
		},
	}
	dbErr := errors.New("unexpected error")

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_GetAllCategories/successful_get_all_categories"
	t.Run("successful_get_all_categories", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().FindAll(mock.Anything).Return(categories, nil)

		res, err := uc.GetAllCategories(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 2)
		assert.Equal(t, categories[0].Name, res[0].Name)
		assert.Equal(t, categories[1].Name, res[1].Name)
	})

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_GetAllCategories/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().FindAll(mock.Anything).Return(nil, dbErr)

		result, err := uc.GetAllCategories(ctx)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists"
func TestCategoryUsecaseImpl_ValidateCategoryExists(t *testing.T) {
	ctx := context.Background()
	categoryID := uint(1)
	category := entity.Category{
		ID:   categoryID,
		Name: "Test Category",
		Slug: "test-category",
	}
	dbErr := errors.New("unexpected error")

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/successful_validate_category_exists"
	t.Run("successful_validate_category_exists", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, categoryID).Return(&category, nil)

		err := uc.ValidateCategoryExists(ctx, categoryID)
		assert.NoError(t, err)
	})

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/failed_to_validate_category_not_found"
	t.Run("failed_to_validate_category_not_found", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, categoryID).Return(nil, apperror.ErrRecordNotFound)

		err := uc.ValidateCategoryExists(ctx, categoryID)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo := setupCategoryUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, categoryID).Return(nil, dbErr)

		err := uc.ValidateCategoryExists(ctx, categoryID)
		assert.ErrorIs(t, err, dbErr)
	})
}
