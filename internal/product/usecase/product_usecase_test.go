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

	//go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_duplicate_slug"
	t.Run("failed_duplicate_slug", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, transactionMock := setupProductUsecase(t)
		request := &dto.ProductCreateRequest{
			CategoryID: 1,
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
		repo.On("Create", mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProduct)
		result, err := usecase.CreateProduct(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProduct)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_duplicate_sku"
	t.Run("failed_duplicate_sku", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, transactionMock := setupProductUsecase(t)
		request := &dto.ProductCreateRequest{
			CategoryID: 1,
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
		repo.On("Create", mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProductSku)
		result, err := usecase.CreateProduct(ctx, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProductSku)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_CreateProduct/failed_unexpected_error_from_category_usecase"
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

func TestProductUsecaseImpl_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	isActiveTrue := true

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/success_update_product"
	t.Run("success_update_product", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, transactionMock := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID:  uint(2),
			Name:        "Product Updated Name",
			Description: "Updated Description",
			Price:       15000,
			Stock:       20,
			SKU:         "PROD-UPDATED-001",
			IsActive:    &isActiveTrue,
		}
		existingProduct := &entity.Product{
			ID:         productID,
			CategoryID: uint(1),
			Name:       "Old Product Name",
			Slug:       "old-product-name",
			Price:      10000,
			Stock:      10,
			SKU:        "PROD-OLD-001",
			IsActive:   false,
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(existingProduct, nil)
		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(nil)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.ID == productID &&
				p.CategoryID == request.CategoryID &&
				p.Name == request.Name &&
				p.Slug == "product-updated-name" &&
				p.SKU == "PROD-UPDATED-001"
		})).Return(nil)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, "product-updated-name", result.Slug)
		assert.Equal(t, "PROD-UPDATED-001", result.SKU)
		assert.True(t, result.IsActive)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_product_not_found"
	t.Run("failed_product_not_found", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		productID := uint(99)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(1),
			Name:       "Test",
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(nil, apperror.ErrNotFound)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_category_not_found"
	t.Run("failed_category_not_found", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, _ := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(3),
			Name:       "Test",
		}
		existingProduct := &entity.Product{
			ID:         productID,
			CategoryID: uint(1),
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(existingProduct, nil)
		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(apperror.ErrCategoryNotFound)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_duplicate_product_slug"
	t.Run("failed_duplicate_product_slug", func(t *testing.T) {
		usecase, repo, _, transactionMock := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(1),
			Name:       "Slug Test",
		}
		existingProduct := &entity.Product{
			ID:         productID,
			CategoryID: uint(1),
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(existingProduct, nil)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("Update", mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProduct)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProduct)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_duplicate_sku"
	t.Run("failed_duplicate_sku", func(t *testing.T) {
		usecase, repo, _, transactionMock := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(1),
			Name:       "Test",
			SKU:        "SKU-TEST",
		}
		existingProduct := &entity.Product{
			ID:         productID,
			CategoryID: uint(1),
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(existingProduct, nil)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("Update", mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProductSku)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProductSku)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(1),
			Name:       "Test",
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(nil, errors.New("error from repository"))
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_UpdateProduct/failed_unexpected_error_from_category"
	t.Run("failed_unexpected_error_from_category", func(t *testing.T) {
		usecase, repo, categoryUsecaseMock, _ := setupProductUsecase(t)
		productID := uint(1)
		request := &dto.ProductUpdateRequest{
			CategoryID: uint(2),
			Name:       "Test",
		}
		existingProduct := &entity.Product{
			ID:         productID,
			CategoryID: uint(1),
		}
		repo.On("FindByID", mock.Anything, productID).
			Return(existingProduct, nil)
		categoryUsecaseMock.On("ValidateCategoryExists", mock.Anything, request.CategoryID).
			Return(apperror.ErrInternalServer)
		result, err := usecase.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

func TestProductUsecaseImpl_DeleteProduct(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DeleteProduct/success_delete_product"
	t.Run("success_delete_product", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("Delete", mock.Anything, productID).Return(nil)
		err := usecase.DeleteProduct(ctx, productID)
		assert.NoError(t, err)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DeleteProduct/failed_product_not_found"
	t.Run("failed_product_not_found", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("Delete", mock.Anything, productID).Return(apperror.ErrNotFound)
		err := usecase.DeleteProduct(ctx, productID)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_DeleteProduct/failed_unexpected_db_error"
	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("Delete", mock.Anything, productID).Return(errors.New("unexpected error"))
		err := usecase.DeleteProduct(ctx, productID)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_SearchProducts"
func TestProductUsecaseImpl_SearchProducts(t *testing.T) {
	ctx := context.Background()

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_SearchProducts/success_search_products_found"
	t.Run("success_search_products_found", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)

		request := &dto.ProductSearchRequest{
			Search:     "Baju",
			CategoryID: uint(1),
			Page:       1,
			Limit:      10,
		}

		mockProducts := []entity.Product{
			{ID: 1, CategoryID: 1, Name: "Baju Koko", Slug: "baju-koko", Price: 50000, Stock: 10, IsActive: true},
			{ID: 2, CategoryID: 1, Name: "Baju Kaos", Slug: "baju-kaos", Price: 35000, Stock: 5, IsActive: true},
		}
		totalData := int64(2)

		repo.On("FindAll", mock.Anything, mock.MatchedBy(func(f *entity.ProductFilter) bool {
			return f.Search == request.Search && f.CategoryID == request.CategoryID && f.Page == request.Page && f.Limit == request.Limit
		})).Return(mockProducts, totalData, nil)

		result, err := usecase.SearchProducts(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, totalData, result.Meta.Total)
		assert.Equal(t, request.Page, result.Meta.Page)
		assert.Equal(t, "Baju Koko", result.Data[0].Name)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_SearchProducts/success_search_products_empty_result"
	t.Run("success_search_products_empty_result", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)

		request := &dto.ProductSearchRequest{
			Search: "BarangGaib",
			Page:   1,
			Limit:  10,
		}

		repo.On("FindAll", mock.Anything, mock.Anything).
			Return([]entity.Product{}, int64(0), nil)

		result, err := usecase.SearchProducts(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 0)
		assert.Equal(t, int64(0), result.Meta.Total)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_SearchProducts/failed_unexpected_db_error"
	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)

		request := &dto.ProductSearchRequest{Page: 1, Limit: 10}

		repo.On("FindAll", mock.Anything, mock.Anything).
			Return(nil, int64(0), errors.New("unexpected error"))

		result, err := usecase.SearchProducts(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductDetail"
func TestProductUsecaseImpl_GetProductDetail(t *testing.T) {
	ctx := context.Background()

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductDetail/successful_get_product_detail"
	t.Run("successful_get_product_detail", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("FindByID", mock.Anything, uint(1)).
			Return(&entity.Product{
				ID:       1,
				Name:     "Test",
				IsActive: true,
			}, nil)
		result, err := usecase.GetProductDetail(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductDetail/failed_product_inactive_treated_as_not_found"
	t.Run("failed_product_inactive_treated_as_not_found", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("FindByID", mock.Anything, uint(1)).
			Return(&entity.Product{
				ID:       1,
				Name:     "Test",
				IsActive: false,
			}, nil)
		result, err := usecase.GetProductDetail(ctx, 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductDetail/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, repo, _, _ := setupProductUsecase(t)
		repo.On("FindByID", mock.Anything, uint(1)).
			Return(nil, errors.New("unexpected error"))
		result, err := usecase.GetProductDetail(ctx, 1)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_AdjustStock"
func TestProductUsecaseImpl_AdjustStock(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	newStock := 50

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_AdjustStock/success_adjust_stock"
	t.Run("success_adjust_stock", func(t *testing.T) {
		usecase, repo, _, transactionMock := setupProductUsecase(t)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("AdjustStock", mock.Anything, productID, newStock).Return(nil)
		err := usecase.AdjustStock(ctx, productID, newStock)
		assert.NoError(t, err)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_AdjustStock/failed_product_not_found"
	t.Run("failed_product_not_found", func(t *testing.T) {
		usecase, repo, _, transactionMock := setupProductUsecase(t)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("AdjustStock", mock.Anything, productID, newStock).Return(apperror.ErrNotFound)
		err := usecase.AdjustStock(ctx, productID, newStock)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_AdjustStock/failed_unexpected_db_error"
	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		usecase, repo, _, transactionMock := setupProductUsecase(t)
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		repo.On("AdjustStock", mock.Anything, productID, newStock).Return(errors.New("unexpected error"))
		err := usecase.AdjustStock(ctx, productID, newStock)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetByProductID"
func TestProductUsecaseImpl_GetByProductID(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetByProductID/success_get_active_product"
	t.Run("success_get_active_product", func(t *testing.T) {
		service, repo := setupProductService(t)

		mockProduct := &entity.Product{ID: productID, Name: "Produk Aktif", IsActive: true}
		repo.On("FindByID", mock.Anything, productID).Return(mockProduct, nil)

		result, err := service.GetByProductID(ctx, productID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsActive)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetByProductID/failed_product_not_found"
	t.Run("failed_product_not_found", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("FindByID", mock.Anything, productID).Return(nil, apperror.ErrNotFound)

		result, err := service.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetByProductID/failed_product_inactive"
	t.Run("failed_product_inactive", func(t *testing.T) {
		service, repo := setupProductService(t)

		mockProduct := &entity.Product{ID: productID, Name: "Produk Mati", IsActive: false}
		repo.On("FindByID", mock.Anything, productID).Return(mockProduct, nil)

		result, err := service.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetByProductID/failed_unexpected_db_error"
	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("FindByID", mock.Anything, productID).Return(nil, errors.New("unexpected error"))

		result, err := service.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductsByIDs"
func TestProductUsecaseImpl_GetProductsByIDs(t *testing.T) {
	ctx := context.Background()
	productIDs := []uint{1, 2, 3}

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductsByIDs/success_get_all_active_products"
	t.Run("success_get_all_active_products", func(t *testing.T) {
		service, repo := setupProductService(t)

		mockProducts := []entity.Product{
			{ID: 1, Name: "Produk 1", IsActive: true},
			{ID: 2, Name: "Produk 2", IsActive: true},
		}

		repo.On("FindByIDs", mock.Anything, productIDs).Return(mockProducts, nil)

		result, err := service.GetProductsByIDs(ctx, productIDs)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductsByIDs/success_but_inactive_products_are_filtered_out"
	t.Run("success_but_inactive_products_are_filtered_out", func(t *testing.T) {
		service, repo := setupProductService(t)

		mockProducts := []entity.Product{
			{ID: 1, Name: "Produk 1 Aktif", IsActive: true},
			{ID: 2, Name: "Produk 2 Mati", IsActive: false},
			{ID: 3, Name: "Produk 3 Aktif", IsActive: true},
		}

		repo.On("FindByIDs", mock.Anything, productIDs).Return(mockProducts, nil)

		result, err := service.GetProductsByIDs(ctx, productIDs)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, uint(1), result[0].ID)
		assert.Equal(t, uint(3), result[1].ID)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductsByIDs/failed_all_retrieved_products_are_inactive"
	t.Run("failed_all_retrieved_products_are_inactive", func(t *testing.T) {
		service, repo := setupProductService(t)

		mockProducts := []entity.Product{
			{ID: 1, Name: "Produk 1 Mati", IsActive: false},
			{ID: 2, Name: "Produk 2 Mati", IsActive: false},
		}

		repo.On("FindByIDs", mock.Anything, productIDs).Return(mockProducts, nil)

		result, err := service.GetProductsByIDs(ctx, productIDs)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	// go test -v ./internal/product/usecase -run "TestProductUsecaseImpl_GetProductsByIDs/failed_unexpected_db_error"
	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		service, repo := setupProductService(t)

		repo.On("FindByIDs", mock.Anything, productIDs).Return(nil, errors.New("unexpected error"))

		result, err := service.GetProductsByIDs(ctx, productIDs)

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
