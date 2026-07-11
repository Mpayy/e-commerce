package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	configMock "github.com/Mpayy/e-commerce/internal/mocks"
	"github.com/Mpayy/e-commerce/internal/product/dto"
	"github.com/Mpayy/e-commerce/internal/product/entity"
	repoMock "github.com/Mpayy/e-commerce/internal/product/mocks"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLoggerCategory() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func setupCategoryUsecase(t *testing.T) (CategoryUsecase, *repoMock.MockCategoryRepository, *configMock.MockTransaction) {
	categoryRepository := repoMock.NewMockCategoryRepository(t)
	transactionMock := configMock.NewMockTransaction(t)
	log := newTestLoggerCategory()
	usecase := NewCategoryUsecase(categoryRepository, log, transactionMock)
	return usecase, categoryRepository, transactionMock
}

// go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory"
func TestCategoryUsecaseImpl_CreateCategory(t *testing.T) {
	ctx := context.Background()

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/successful_create_category"
	t.Run("successful_create_category", func(t *testing.T) {
		usecase, repo, transactionMock := setupCategoryUsecase(t)

		request := &dto.CategoryRequest{
			Name: "Test Category",
		}

		requestSlug := slug.Make(request.Name)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			if u.Name != request.Name || u.Slug != requestSlug {
				return false
			}
			return true
		})).
			Return(nil)

		result, err := usecase.CreateCategory(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, requestSlug, result.Slug)
	})

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/failed_to_create_category"
	t.Run("failed_to_create_category", func(t *testing.T) {
		usecase, repo, transactionMock := setupCategoryUsecase(t)

		request := &dto.CategoryRequest{
			Name: "Test Category",
		}

		requestSlug := slug.Make(request.Name)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			if u.Name != request.Name || u.Slug != requestSlug {
				return false
			}
			return true
		})).
			Return(apperror.ErrDuplicatedKey)

		result, err := usecase.CreateCategory(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedCategory)
	})

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_CreateCategory/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, repo, transactionMock := setupCategoryUsecase(t)

		request := &dto.CategoryRequest{
			Name: "Test Category",
		}

		requestSlug := slug.Make(request.Name)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.Category) bool {
			if u.Name != request.Name || u.Slug != requestSlug {
				return false
			}
			return true
		})).
			Return(errors.New("unexpected error"))

		result, err := usecase.CreateCategory(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

func TestCategoryUsecaseImpl_ValidateCategoryExists(t *testing.T) {
	ctx := context.Background()

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/successful_validate_category_exists"
	t.Run("successful_validate_category_exists", func(t *testing.T) {
		usecase, repo, _ := setupCategoryUsecase(t)

		categoryID := uint(1)

		repo.On("FindByID", mock.Anything, categoryID).
			Return(&entity.Category{
				ID:   categoryID,
				Name: "Test Category",
				Slug: "test-category",
			}, nil)

		err := usecase.ValidateCategoryExists(ctx, categoryID)
		assert.NoError(t, err)
	})

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/failed_to_validate_category_exists"
	t.Run("failed_to_validate_category_exists", func(t *testing.T) {
		usecase, repo, _ := setupCategoryUsecase(t)

		categoryID := uint(1)

		repo.On("FindByID", mock.Anything, categoryID).
			Return(nil, apperror.ErrNotFound)

		err := usecase.ValidateCategoryExists(ctx, categoryID)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	//go test -v ./internal/product/usecase -run "TestCategoryUsecaseImpl_ValidateCategoryExists/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, repo, _ := setupCategoryUsecase(t)

		categoryID := uint(1)

		repo.On("FindByID", mock.Anything, categoryID).
			Return(nil, errors.New("unexpected error"))

		err := usecase.ValidateCategoryExists(ctx, categoryID)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}
	