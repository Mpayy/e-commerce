package usecase

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/internal/product/dto"
	"github.com/Mpayy/e-commerce/internal/product/entity"
	productrepository "github.com/Mpayy/e-commerce/internal/product/repository"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/skugen"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

type ProductUsecaseImpl struct {
	productRepository productrepository.ProductRepository
	categoryUsecase   CategoryUsecase
	log               *logrus.Logger
	transaction       transaction.Transaction
}

func NewProductUsecase(productRepository productrepository.ProductRepository, categoryUsecase CategoryUsecase, log *logrus.Logger, transaction transaction.Transaction) *ProductUsecaseImpl {
	return &ProductUsecaseImpl{
		productRepository: productRepository,
		categoryUsecase:   categoryUsecase,
		log:               log,
		transaction:       transaction,
	}
}

func (u *ProductUsecaseImpl) CreateProduct(ctx context.Context, request *dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	logger := u.log.WithField("name", request.Name)
	logger.Debug("Attempting to create product")

	err := u.categoryUsecase.ValidateCategoryExists(ctx, request.CategoryID)
	if err != nil {
		return nil, err
	}

	sku := skugen.Sanitize(request.SKU)
	if sku == "" {
		sku = skugen.Generate()
	}

	product := &entity.Product{
		CategoryID:  request.CategoryID,
		Name:        request.Name,
		Slug:        slug.Make(request.Name),
		Description: request.Description,
		Price:       request.Price,
		Stock:       request.Stock,
		SKU:         sku,
		IsActive:    true,
	}

	if err = u.productRepository.Create(ctx, product); err != nil {
		if errors.Is(err, apperror.ErrDuplicatedProduct) {
			logger.WithField("slug", product.Slug).Warn("Create product failed: duplicate slug")
			return nil, err
		}
		if errors.Is(err, apperror.ErrDuplicatedProductSku) {
			logger.WithField("sku", product.SKU).Warn("Create product failed: duplicate SKU")
			return nil, err
		}
		logger.WithError(err).Error("Failed to create product")
		return nil, apperror.ErrInternalServer
	}

	response := &dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
	}

	logger.Info("Product created successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) UpdateProduct(ctx context.Context, id uint, request *dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	logger := u.log.WithField("id", id)
	logger.Debug("Attempting to update product")

	product, err := u.productRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to update product: product not found")
			return nil, apperror.ErrProductNotFound
		}
		logger.WithError(err).Error("Failed to update product")
		return nil, apperror.ErrInternalServer
	}

	if product.CategoryID != request.CategoryID {
		err = u.categoryUsecase.ValidateCategoryExists(ctx, request.CategoryID)
		if err != nil {
			return nil, err
		}
	}

	product.CategoryID = request.CategoryID
	product.Name = request.Name
	product.Slug = slug.Make(request.Name)
	product.Description = request.Description
	product.Price = request.Price
	product.Stock = request.Stock

	if request.SKU != "" {
		product.SKU = skugen.Sanitize(request.SKU)
	}

	if request.IsActive != nil {
		product.IsActive = *request.IsActive
	}

	if err := u.productRepository.Update(ctx, product); err != nil {
		if errors.Is(err, apperror.ErrDuplicatedProduct) {
			logger.WithField("slug", product.Slug).Warn("Update product failed: duplicate slug")
			return nil, err
		}

		if errors.Is(err, apperror.ErrDuplicatedProductSku) {
			logger.WithField("sku", product.SKU).Warn("Update product failed: duplicate SKU")
			return nil, err
		}
		logger.WithError(err).Error("Failed to update product")
		return nil, apperror.ErrInternalServer
	}

	response := &dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
	}

	logger.Info("Product updated successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) DeleteProduct(ctx context.Context, id uint) error {
	logger := u.log.WithField("id", id)
	logger.Debug("Attempting to delete product")

	err := u.productRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to delete product: product not found")
			return apperror.ErrProductNotFound
		}
		logger.WithError(err).Error("Failed to delete product: unexpected DB error")
		return apperror.ErrInternalServer
	}

	logger.Info("Product deleted successfully")
	return nil
}

func (u *ProductUsecaseImpl) SearchProducts(ctx context.Context, request *dto.ProductSearchRequest) (*dto.ProductSearchResponse, error) {
	logger := u.log.WithFields(logrus.Fields{
		"search":      request.Search,
		"category_id": request.CategoryID,
		"page":        request.Page,
		"limit":       request.Limit,
	})
	logger.Debug("Attempting to search products")

	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit <= 0 {
		request.Limit = 10
	}

	filter := &entity.ProductFilter{
		Search:     request.Search,
		CategoryID: request.CategoryID,
		Page:       request.Page,
		Limit:      request.Limit,
	}

	products, total, err := u.productRepository.FindAll(ctx, filter)
	if err != nil {
		logger.WithError(err).Error("Failed to search products")
		return nil, apperror.ErrInternalServer
	}

	response := []dto.ProductResponse{}
	for _, product := range products {
		response = append(response, dto.ProductResponse{
			ID:          product.ID,
			CategoryID:  product.CategoryID,
			Name:        product.Name,
			Slug:        product.Slug,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			SKU:         product.SKU,
			IsActive:    product.IsActive,
		})
	}

	logger.Info("Products searched successfully")
	return &dto.ProductSearchResponse{
		Data: response,
		Meta: dto.MetaPagination{
			Total: total,
			Page:  filter.Page,
			Limit: filter.Limit,
		},
	}, nil

}

func (u *ProductUsecaseImpl) GetProductDetail(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	logger := u.log.WithField("id", id)
	logger.Debug("Attempting to get product detail")

	product, err := u.productRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to get product detail: product not found")
			return nil, apperror.ErrProductNotFound
		}

		logger.WithError(err).Error("Failed to get product detail: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	if !product.IsActive {
		logger.Warn("Get product detail failed: product not active")
		return nil, apperror.ErrProductNotFound
	}

	response := &dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
	}

	logger.Info("Product detail retrieved successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) AdjustStock(ctx context.Context, productID uint, stock int) error {
	logger := u.log.WithFields(logrus.Fields{
		"product_id": productID,
		"stock":      stock,
	})
	logger.Debug("Attempting to adjust stock")

	err := u.productRepository.AdjustStock(ctx, productID, stock)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to adjust stock: product not found")
			return apperror.ErrProductNotFound
		}

		logger.WithError(err).Error("Failed to adjust stock: unexpected DB error")
		return apperror.ErrInternalServer
	}

	logger.Info("Stock adjusted successfully")
	return nil
}

// ═══════════════════════════════════════════════════════
// Consumption By Other Services (contract.go)
// ═══════════════════════════════════════════════════════
func (u *ProductUsecaseImpl) GetByProductID(ctx context.Context, id uint) (*entity.Product, error) {
	logger := u.log.WithField("id", id)
	logger.Debug("Attempting to get product by ID")

	product, err := u.productRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to get product: product not found")
			return nil, apperror.ErrProductNotFound
		}

		logger.WithError(err).Error("Failed to get product: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	if !product.IsActive {
		logger.Warn("Get product failed: product not active")
		return nil, apperror.ErrProductNotFound
	}
	logger.Debug("Product retrieved")
	return product, nil
}

func (u *ProductUsecaseImpl) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	logger := u.log.WithField("ids", ids)
	logger.Debug("Attempting to get products by IDs")

	products, err := u.productRepository.FindByIDs(ctx, ids)
	if err != nil {
		logger.WithError(err).Error("Failed to get products: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	var result []entity.Product
	for _, product := range products {
		if !product.IsActive {
			continue
		}
		result = append(result, product)
	}
	if len(result) == 0 {
		logger.Warn("Get products failed: products not found")
		return nil, apperror.ErrProductNotFound
	}

	logger.Debug("Products retrieved")
	return result, nil
}

func (u *ProductUsecaseImpl) DecreaseStock(ctx context.Context, productID uint, quantity int) error {
	logger := u.log.WithFields(logrus.Fields{
		"product_id": productID,
		"quantity":   quantity,
	})
	logger.Debug("Attempting to decrease stock")
	err := u.productRepository.DecreaseStock(ctx, productID, quantity)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			logger.Warn("Failed to decrease stock: product not found")
			return apperror.ErrProductNotFound
		}
		if errors.Is(err, apperror.ErrInsufficientStock) {
			logger.Warn("Failed to decrease stock: insufficient stock")
			return err
		}
		logger.WithError(err).Error("Failed to decrease stock: unexpected DB error")
		return apperror.ErrInternalServer
	}
	logger.Info("Stock decreased successfully")
	return nil
}
