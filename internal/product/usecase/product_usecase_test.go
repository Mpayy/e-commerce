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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLoggerProduct() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func setupProductUsecase(t *testing.T) (ProductUsecase, *repoMock.MockProductRepository, *repoMock.MockCategoryUsecase, *configMock.MockTransaction) {
	log := newTestLoggerProduct()
	transactionMock := configMock.NewMockTransaction(t)
	productRepository := repoMock.NewMockProductRepository(t)

	categoryUsecase := repoMock.NewMockCategoryUsecase(t)
	productUsecase := NewProductUsecase(productRepository, categoryUsecase, log, transactionMock)

	return productUsecase, productRepository, categoryUsecase, transactionMock
}

func setupProductService(t *testing.T) (ProductService, *repoMock.MockProductRepository) {
	productRepository := repoMock.NewMockProductRepository(t)
	log := newTestLoggerProduct()
	productService := NewProductUsecase(productRepository, nil, log, nil)

	return productService, productRepository
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct"
func TestProductUsecaseImpl_CreateProduct(t *testing.T) {
	ctx := context.Background()

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/successful_create_product"
	t.Run("successful_create_product", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, transactionMock := setupProductUsecase(t)

		request := &dto.ProductCreateRequest{
			CategoryID: uint(1),
			Name:       "Test Product",
			Price:      10000,
			Stock:      10,
		}

		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(nil)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			if p.CategoryID != request.CategoryID || p.Name != request.Name || p.Price != request.Price || p.Stock != request.Stock {
				return false
			}
			if p.SKU == "" {
				return false
			}
			return true
		})).Return(nil)

		result, err := usecase.CreateProduct(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, request.Price, result.Price)
		assert.Equal(t, request.Stock, result.Stock)
		assert.Equal(t, request.CategoryID, result.CategoryID)
		assert.Contains(t, result.SKU, "PRD-")
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_create_product_category_not_found"
	t.Run("failed_create_product_category_not_found", func(t *testing.T) {
		usecase, _, categoryUsecaseMock, _ := setupProductUsecase(t)

		request := &dto.ProductCreateRequest{
			CategoryID: uint(1),
			Name:       "Test Product",
			Price:      10000,
			Stock:      10,
		}

		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(apperror.ErrCategoryNotFound)

		result, err := usecase.CreateProduct(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_unexpected_error_from_category_usecase"
	t.Run("failed_unexpected_error_from_category_usecase", func(t *testing.T) {
		usecase, _, categoryUsecaseMock, _ := setupProductUsecase(t)

		request := &dto.ProductCreateRequest{
			CategoryID: uint(1),
			Name:       "Test Product",
			Price:      10000,
			Stock:      10,
		}

		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(apperror.ErrInternalServer)

		result, err := usecase.CreateProduct(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, transactionMock := setupProductUsecase(t)

		request := &dto.ProductCreateRequest{
			CategoryID: uint(1),
			Name:       "Test Product",
			Price:      10000,
			Stock:      10,
		}

		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(nil)

		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			if p.CategoryID != request.CategoryID || p.Name != request.Name || p.Price != request.Price || p.Stock != request.Stock {
				return false
			}
			if p.SKU == "" {
				return false
			}
			return true
		})).Return(errors.New("unexpected error"))

		result, err := usecase.CreateProduct(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})

}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DecreaseStock"
func TestProductUsecaseImpl_DecreaseStock(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	quantity := 10

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DecreaseStock/successful_decrease_stock"
	t.Run("successful_decrease_stock", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("DecreaseStock", mock.Anything, productID, quantity).Return(nil)

		err := service.DecreaseStock(ctx, productID, quantity)
		assert.NoError(t, err)
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DecreaseStock/failed_product_not_found"
	t.Run("failed_product_not_found", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("DecreaseStock", mock.Anything, productID, quantity).
			Return(apperror.ErrNotFound)

		err := service.DecreaseStock(ctx, productID, quantity)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DecreaseStock/failed_insufficient_stock"
	t.Run("failed_insufficient_stock", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("DecreaseStock", mock.Anything, productID, quantity).
			Return(apperror.ErrInsufficientStock)

		err := service.DecreaseStock(ctx, productID, quantity)
		assert.ErrorIs(t, err, apperror.ErrInsufficientStock)
	})

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DecreaseStock/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("DecreaseStock", mock.Anything, productID, quantity).
			Return(errors.New("unexpected error"))

		err := service.DecreaseStock(ctx, productID, quantity)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}
