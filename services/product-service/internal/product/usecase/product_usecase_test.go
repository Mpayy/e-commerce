package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/dto"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLoggerProduct() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupProductUsecase(t *testing.T) (ProductUsecase, *mocks.MockProductRepository, *mocks.MockCategoryUsecase) {
	log := newTestLoggerProduct()
	productRepository := mocks.NewMockProductRepository(t)
	categoryUsecase := mocks.NewMockCategoryUsecase(t)
	productUsecase := NewProductUsecase(productRepository, categoryUsecase, log)
	t.Cleanup(func() {
		productRepository.AssertExpectations(t)
		categoryUsecase.AssertExpectations(t)
	})
	return productUsecase, productRepository, categoryUsecase
}

func setupProductService(t *testing.T) (ProductService, *mocks.MockProductRepository) {
	productRepository := mocks.NewMockProductRepository(t)
	log := newTestLoggerProduct()
	productService := NewProductUsecase(productRepository, nil, log)
	t.Cleanup(func() {
		productRepository.AssertExpectations(t)
	})
	return productService, productRepository
}

func TestProductUsecaseImpl_CreateProduct(t *testing.T) {
	ctx := context.Background()
	request := &dto.ProductCreateRequest{
		CategoryID: uint(1),
		Name:       "Test Product",
		Price:      10000,
		Stock:      10,
	}
	dbErr := errors.New("unexpected error")

	t.Run("successful_create_product", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p != nil &&
				p.CategoryID == request.CategoryID &&
				p.Name == request.Name &&
				p.Price == request.Price &&
				p.Stock == request.Stock &&
				p.SKU != ""
		})).Return(nil)

		result, err := uc.CreateProduct(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, request.Price, result.Price)
		assert.Equal(t, request.Stock, result.Stock)
		assert.Equal(t, request.CategoryID, result.CategoryID)
		assert.Contains(t, result.SKU, "PRD-")
	})

	t.Run("failed_create_product_category_not_found", func(t *testing.T) {
		uc, _, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(apperror.ErrCategoryNotFound)

		result, err := uc.CreateProduct(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	t.Run("failed_duplicate_slug", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(nil)

		repo.On("Create", mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProduct)

		result, err := uc.CreateProduct(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProduct)
	})

	t.Run("failed_duplicate_sku", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(nil)

		repo.EXPECT().Create(mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProductSku)

		result, err := uc.CreateProduct(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProductSku)
	})

	t.Run("failed_unexpected_error_from_category_usecase", func(t *testing.T) {
		uc, _, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(dbErr)

		result, err := uc.CreateProduct(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, request.CategoryID).
			Return(nil)

		repo.EXPECT().Create(mock.Anything, mock.Anything).
			Return(dbErr)

		result, err := uc.CreateProduct(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

}

func TestProductUsecaseImpl_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	isActiveTrue := true
	productID := uint(1)
	categoryID := uint(2)

	request := &dto.ProductUpdateRequest{
		CategoryID:  categoryID,
		Name:        "Product Updated Name",
		Description: "Updated Description",
		Price:       15000,
		Stock:       20,
		SKU:         "PROD-UPDATED-001",
		IsActive:    &isActiveTrue,
	}

	getFreshProduct := func() *entity.Product {
		return &entity.Product{
			ID:          productID,
			CategoryID:  uint(1),
			Name:        "Old Product Name",
			Slug:        "old-product-name",
			Description: "Old Description",
			Price:       10000,
			Stock:       10,
			SKU:         "PROD-OLD-001",
			IsActive:    false,
		}
	}

	dbErr := errors.New("unexpected error")

	t.Run("success_update_product_category_changed", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(nil)

		repo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.ID == productID &&
				p.CategoryID == categoryID &&
				p.Name == request.Name &&
				p.Slug == "product-updated-name" &&
				p.SKU == "PROD-UPDATED-001" &&
				p.IsActive == true
		})).Return(nil)

		result, err := uc.UpdateProduct(ctx, productID, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, "product-updated-name", result.Slug)
		assert.Equal(t, "PROD-UPDATED-001", result.SKU)
		assert.True(t, result.IsActive)
	})

	t.Run("success_update_product_same_category_and_nil_optional_fields", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)
		product := getFreshProduct()

		reqSameCategory := &dto.ProductUpdateRequest{
			CategoryID:  product.CategoryID,
			Name:        "Updated Name Same Category",
			Description: "Updated Desc",
			Price:       20000,
			Stock:       50,
			SKU:         "",
			IsActive:    nil,
		}

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		// categoryUC.ValidateCategoryExists TIDAK dipanggil di sini

		repo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.CategoryID == product.CategoryID &&
				p.SKU == "PROD-OLD-001" &&
				p.IsActive == false
		})).Return(nil)

		result, err := uc.UpdateProduct(ctx, productID, reqSameCategory)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, reqSameCategory.Name, result.Name)
		assert.Equal(t, "PROD-OLD-001", result.SKU)
		assert.False(t, result.IsActive)
	})

	t.Run("failed_product_not_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(nil, apperror.ErrRecordNotFound)

		result, err := uc.UpdateProduct(ctx, productID, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_error_from_find_by_id", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(nil, dbErr)

		result, err := uc.UpdateProduct(ctx, productID, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("failed_category_not_found", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(apperror.ErrCategoryNotFound)

		result, err := uc.UpdateProduct(ctx, productID, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrCategoryNotFound)
	})

	t.Run("failed_unexpected_error_from_category", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(dbErr)

		result, err := uc.UpdateProduct(ctx, productID, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("failed_duplicate_product_slug", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(nil)

		repo.EXPECT().Update(mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProduct)

		result, err := uc.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProduct)
	})

	t.Run("failed_duplicate_sku", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(nil)

		repo.EXPECT().Update(mock.Anything, mock.Anything).
			Return(apperror.ErrDuplicatedProductSku)

		result, err := uc.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedProductSku)
	})

	t.Run("failed_unexpected_error_from_update_repository", func(t *testing.T) {
		uc, repo, categoryUC := setupProductUsecase(t)
		product := getFreshProduct()

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		categoryUC.EXPECT().ValidateCategoryExists(mock.Anything, categoryID).
			Return(nil)

		repo.EXPECT().Update(mock.Anything, mock.Anything).
			Return(dbErr)

		result, err := uc.UpdateProduct(ctx, productID, request)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_DeleteProduct(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	dbErr := errors.New("unexpected error")

	t.Run("success_delete_product", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().Delete(mock.Anything, productID).
			Return(nil)

		err := uc.DeleteProduct(ctx, productID)
		assert.NoError(t, err)
	})

	t.Run("failed_product_not_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().Delete(mock.Anything, productID).
			Return(apperror.ErrRecordNotFound)

		err := uc.DeleteProduct(ctx, productID)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().Delete(mock.Anything, productID).
			Return(dbErr)

		err := uc.DeleteProduct(ctx, productID)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_SearchProducts(t *testing.T) {
	ctx := context.Background()

	products := []entity.Product{
		{
			ID:          1,
			CategoryID:  1,
			Name:        "Baju Koko",
			Slug:        "baju-koko",
			Description: "Deskripsi Baju Koko",
			Price:       50000,
			Stock:       10,
			SKU:         "SKU-KOKO-001",
			IsActive:    true,
		},
		{
			ID:          2,
			CategoryID:  1,
			Name:        "Baju Kaos",
			Slug:        "baju-kaos",
			Description: "Deskripsi Baju Kaos",
			Price:       35000,
			Stock:       5,
			SKU:         "SKU-KAOS-001",
			IsActive:    true,
		},
	}
	totalData := int64(2)
	dbErr := errors.New("unexpected error")

	t.Run("success_search_products_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		req := &dto.ProductSearchRequest{
			Search:     "Baju",
			CategoryID: uint(1),
			Page:       1,
			Limit:      10,
		}

		repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f *entity.ProductFilter) bool {
			return f != nil &&
				f.Search == req.Search &&
				f.CategoryID == req.CategoryID &&
				f.Page == req.Page &&
				f.Limit == req.Limit
		})).Return(products, totalData, nil)

		result, err := uc.SearchProducts(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, totalData, result.Meta.Total)
		assert.Equal(t, req.Page, result.Meta.Page)
		assert.Equal(t, req.Limit, result.Meta.Limit)
		assert.Equal(t, "Baju Koko", result.Data[0].Name)
		assert.Equal(t, "baju-koko", result.Data[0].Slug)
		assert.Equal(t, "SKU-KOKO-001", result.Data[0].SKU)
	})

	t.Run("success_search_products_with_default_page_and_limit", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		reqInvalidPagination := &dto.ProductSearchRequest{
			Search:     "Baju",
			CategoryID: uint(1),
			Page:       0,
			Limit:      -5,
		}

		repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f *entity.ProductFilter) bool {
			return f != nil &&
				f.Page == 1 &&
				f.Limit == 10
		})).Return(products, totalData, nil)

		result, err := uc.SearchProducts(ctx, reqInvalidPagination)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Meta.Page)
		assert.Equal(t, 10, result.Meta.Limit)
	})

	t.Run("success_search_products_empty_result", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		req := &dto.ProductSearchRequest{
			Search: "Tidak Ada",
			Page:   1,
			Limit:  10,
		}

		repo.EXPECT().FindAll(mock.Anything, mock.Anything).
			Return([]entity.Product{}, int64(0), nil)

		result, err := uc.SearchProducts(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 0)
		assert.Equal(t, int64(0), result.Meta.Total)
	})

	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		req := &dto.ProductSearchRequest{
			Search: "Baju",
			Page:   1,
			Limit:  10,
		}

		repo.EXPECT().FindAll(mock.Anything, mock.Anything).
			Return(nil, int64(0), dbErr)

		result, err := uc.SearchProducts(ctx, req)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_GetProductDetail(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	productInactiveID := uint(2)

	getFreshProduct := func(active bool) *entity.Product {
		return &entity.Product{
			ID:          productID,
			CategoryID:  1,
			Name:        "Test Product",
			Slug:        "test-product",
			Description: "Test Description",
			Price:       10000,
			Stock:       10,
			SKU:         "SKU-TEST-001",
			IsActive:    active,
		}
	}

	dbErr := errors.New("unexpected error")

	t.Run("successful_get_product_detail", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)
		product := getFreshProduct(true)

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(product, nil)

		res, err := uc.GetProductDetail(ctx, productID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, product.ID, res.ID)
		assert.Equal(t, product.CategoryID, res.CategoryID)
		assert.Equal(t, product.Name, res.Name)
		assert.Equal(t, product.Slug, res.Slug)
		assert.Equal(t, product.Description, res.Description)
		assert.Equal(t, product.Price, res.Price)
		assert.Equal(t, product.Stock, res.Stock)
		assert.Equal(t, product.SKU, res.SKU)
		assert.Equal(t, product.IsActive, res.IsActive)
	})

	t.Run("failed_product_record_not_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(nil, apperror.ErrRecordNotFound)

		res, err := uc.GetProductDetail(ctx, productID)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_product_inactive_treated_as_not_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)
		productInactive := getFreshProduct(false)
		productInactive.ID = productInactiveID

		repo.EXPECT().FindByID(mock.Anything, productInactiveID).
			Return(productInactive, nil)

		res, err := uc.GetProductDetail(ctx, productInactiveID)

		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().FindByID(mock.Anything, productID).
			Return(nil, dbErr)

		result, err := uc.GetProductDetail(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_AdjustStock(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	newStock := 50
	dbErr := errors.New("unexpected error")

	t.Run("success_adjust_stock", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().AdjustStock(mock.Anything, productID, newStock).
			Return(nil)

		err := uc.AdjustStock(ctx, productID, newStock)
		assert.NoError(t, err)
	})

	t.Run("failed_product_not_found", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().AdjustStock(mock.Anything, productID, newStock).
			Return(apperror.ErrRecordNotFound)

		err := uc.AdjustStock(ctx, productID, newStock)

		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		uc, repo, _ := setupProductUsecase(t)

		repo.EXPECT().AdjustStock(mock.Anything, productID, newStock).
			Return(dbErr)

		err := uc.AdjustStock(ctx, productID, newStock)

		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_GetByProductID(t *testing.T) {
	ctx := context.Background()
	productID := uint(1)
	product := &entity.Product{
		ID:       productID,
		Name:     "Produk Aktif",
		IsActive: true,
	}
	productInactive := &entity.Product{
		ID:       productID,
		Name:     "Produk Nonaktif",
		IsActive: false,
	}
	dbErr := errors.New("unexpected error")

	t.Run("success_get_active_product", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByID(mock.Anything, productID).Return(product, nil)

		result, err := srv.GetByProductID(ctx, productID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsActive)
	})

	t.Run("failed_product_not_found", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByID(mock.Anything, productID).Return(nil, apperror.ErrRecordNotFound)

		result, err := srv.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_product_inactive", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByID(mock.Anything, productID).Return(productInactive, nil)

		result, err := srv.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByID(mock.Anything, productID).Return(nil, dbErr)

		result, err := srv.GetByProductID(ctx, productID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_GetProductsByIDs(t *testing.T) {
	ctx := context.Background()
	productIDs := []uint{1, 2, 3}

	productsAllActive := []entity.Product{
		{ID: 1, Name: "Produk 1", IsActive: true},
		{ID: 2, Name: "Produk 2", IsActive: true},
	}
	productsWithInactive := []entity.Product{
		{ID: 1, Name: "Produk 1", IsActive: true},
		{ID: 2, Name: "Produk 2", IsActive: false},
		{ID: 3, Name: "Produk 3", IsActive: true},
	}
	productsAllInactive := []entity.Product{
		{ID: 1, Name: "Produk 1", IsActive: false},
		{ID: 2, Name: "Produk 2", IsActive: false},
	}
	dbErr := errors.New("unexpected error")

	t.Run("success_get_all_active_products", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByIDs(mock.Anything, productIDs).Return(productsAllActive, nil)

		res, err := srv.GetProductsByIDs(ctx, productIDs)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 2)
	})

	t.Run("success_but_inactive_products_are_filtered_out", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByIDs(mock.Anything, productIDs).Return(productsWithInactive, nil)

		res, err := srv.GetProductsByIDs(ctx, productIDs)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, uint(1), res[0].ID)
		assert.Equal(t, uint(3), res[1].ID)
	})

	t.Run("failed_all_retrieved_products_are_inactive", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByIDs(mock.Anything, productIDs).Return(productsAllInactive, nil)

		result, err := srv.GetProductsByIDs(ctx, productIDs)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_unexpected_db_error", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().FindByIDs(mock.Anything, productIDs).Return(nil, dbErr)

		result, err := srv.GetProductsByIDs(ctx, productIDs)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_BulkDecreaseStock(t *testing.T) {
	ctx := context.Background()
	checkoutID := "test-checkout-123"
	items := []entity.BulkDecreaseStock{
		{
			ProductID: 1,
			Quantity:  2,
		},
		{
			ProductID: 2,
			Quantity:  5,
		},
	}
	dbErr := errors.New("unexpected database error")

	t.Run("successful_bulk_decrease_stock", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkDecreaseStock(mock.Anything, checkoutID, items).
			Return(nil)

		err := srv.BulkDecreaseStock(ctx, checkoutID, items)
		assert.NoError(t, err)
	})

	t.Run("failed_product_not_found", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkDecreaseStock(mock.Anything, checkoutID, items).
			Return(apperror.ErrRecordNotFound)

		err := srv.BulkDecreaseStock(ctx, checkoutID, items)

		assert.ErrorIs(t, err, apperror.ErrProductNotFound)
	})

	t.Run("failed_insufficient_stock", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkDecreaseStock(mock.Anything, checkoutID, items).
			Return(apperror.ErrInsufficientStock)

		err := srv.BulkDecreaseStock(ctx, checkoutID, items)

		assert.ErrorIs(t, err, apperror.ErrInsufficientStock)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkDecreaseStock(mock.Anything, checkoutID, items).
			Return(dbErr)

		err := srv.BulkDecreaseStock(ctx, checkoutID, items)

		assert.ErrorIs(t, err, dbErr)
	})
}

func TestProductUsecaseImpl_BulkRestoreStock(t *testing.T) {
	ctx := context.Background()
	checkoutID := "test-checkout-123"
	dbErr := errors.New("unexpected database error")

	t.Run("successful_bulk_restore_stock", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkRestoreStock(mock.Anything, checkoutID).
			Return(nil)

		err := srv.BulkRestoreStock(ctx, checkoutID)
		assert.NoError(t, err)
	})

	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		srv, repo := setupProductService(t)

		repo.EXPECT().
			BulkRestoreStock(mock.Anything, checkoutID).
			Return(dbErr)

		err := srv.BulkRestoreStock(ctx, checkoutID)

		assert.ErrorIs(t, err, dbErr)
	})
}