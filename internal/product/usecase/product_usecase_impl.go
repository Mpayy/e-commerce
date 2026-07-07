package productusecase

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
	ProductRepository productrepository.ProductRepository
	CategoryUsecase   CategoryUsecase
	Log               *logrus.Logger
	Transaction       transaction.Transaction
}

func NewProductUsecase(productRepository productrepository.ProductRepository, categoryUsecase CategoryUsecase, log *logrus.Logger, transaction transaction.Transaction) *ProductUsecaseImpl {
	return &ProductUsecaseImpl{
		ProductRepository: productRepository,
		CategoryUsecase:   categoryUsecase,
		Log:               log,
		Transaction:       transaction,
	}
}

func (u *ProductUsecaseImpl) CreateProduct(ctx context.Context, request *dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	u.Log.WithField("name", request.Name).Debug("Attempting to create product")

	err := u.CategoryUsecase.ValidateCategoryExists(ctx, request.CategoryID)
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

	err = u.Transaction.WithTransaction(ctx, func(ctx context.Context) error {
		errCreate := u.ProductRepository.Create(ctx, product)
		if errCreate != nil {
			if errors.Is(errCreate, apperror.ErrDuplicatedProduct) {
				u.Log.WithField("slug", product.Slug).Warn("Create product failed: duplicate slug")
				return errCreate
			}

			if errors.Is(errCreate, apperror.ErrDuplicatedProductSku) {
				u.Log.WithField("sku", product.SKU).Warn("Create product failed: duplicate SKU")
				return errCreate
			}

			u.Log.WithFields(logrus.Fields{
				"name":  request.Name,
				"error": errCreate,
			}).Error("Create product failed: unexpected DB error")

			return apperror.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
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

	u.Log.WithField("name", request.Name).Info("Product created successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) UpdateProduct(ctx context.Context, id uint, request *dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	u.Log.WithField("id", id).Debug("Attempting to update product")

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("id", id).Warn("Update product failed: product not found")
			return nil, apperror.ErrProductNotFound
		}

		u.Log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Update product failed: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	if product.CategoryID != request.CategoryID {
		err = u.CategoryUsecase.ValidateCategoryExists(ctx, request.CategoryID)
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

	err = u.Transaction.WithTransaction(ctx, func(ctx context.Context) error {
		errUpdate := u.ProductRepository.Update(ctx, product)
		if errUpdate != nil {
			if errors.Is(errUpdate, apperror.ErrDuplicatedProduct) {
				u.Log.WithField("slug", product.Slug).Warn("Update product failed: duplicate slug")
				return errUpdate
			}

			if errors.Is(errUpdate, apperror.ErrDuplicatedProductSku) {
				u.Log.WithField("sku", product.SKU).Warn("Update product failed: duplicate SKU")
				return errUpdate
			}

			u.Log.WithFields(logrus.Fields{
				"id":    id,
				"error": errUpdate,
			}).Error("Update product failed: unexpected DB error")

			return apperror.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
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

	u.Log.WithField("name", request.Name).Info("Product updated successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) DeleteProduct(ctx context.Context, id uint) error {
	u.Log.WithField("id", id).Debug("Attempting to delete product")

	err := u.ProductRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("id", id).Warn("Delete product failed: product not found")
			return apperror.ErrProductNotFound
		}

		u.Log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Delete product failed: unexpected DB error")
		return apperror.ErrInternalServer
	}

	u.Log.WithField("id", id).Info("Product deleted successfully")
	return nil
}

func (u *ProductUsecaseImpl) SearchProducts(ctx context.Context, request *dto.ProductSearchRequest) (*dto.ProductSearchResponse, error) {
	u.Log.WithFields(logrus.Fields{
		"search":      request.Search,
		"category_id": request.CategoryID,
		"page":        request.Page,
		"limit":       request.Limit,
	}).Debug("Attempting to search products")

	filter := &entity.ProductFilter{
		Search:     request.Search,
		CategoryID: request.CategoryID,
		Page:       request.Page,
		Limit:      request.Limit,
	}

	products, total, err := u.ProductRepository.FindAll(ctx, filter)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"filter": filter,
			"error":  err,
		}).Error("Search products failed: unexpected DB error")
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
	u.Log.WithField("id", id).Debug("Attempting to get product detail")

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("id", id).Warn("Get product detail failed: product not found")
			return nil, apperror.ErrProductNotFound
		}

		u.Log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Get product detail failed: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	if !product.IsActive {
		u.Log.WithField("id", id).Warn("Get product detail failed: product not active")
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

	u.Log.WithField("name", response.Name).Info("Product detail retrieved successfully")
	return response, nil
}

func (u *ProductUsecaseImpl) AdjustStock(ctx context.Context, productID uint, stock int) error {
	u.Log.WithFields(logrus.Fields{
		"product_id": productID,
		"stock":      stock,
	}).Debug("Attempting to adjust stock")

	err := u.Transaction.WithTransaction(ctx, func(ctx context.Context) error {
		err := u.ProductRepository.AdjustStock(ctx, productID, stock)
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				u.Log.WithField("product_id", productID).Warn("Adjust stock failed: product not found")
				return apperror.ErrProductNotFound
			}

			u.Log.WithFields(logrus.Fields{
				"product_id": productID,
				"error":      err,
			}).Error("Adjust stock failed: unexpected DB error")
			return apperror.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return err
	}

	u.Log.WithField("product_id", productID).Info("Stock adjusted successfully")
	return nil
}

// ═══════════════════════════════════════════════════════
// Consumption By Other Services (contract.go)
// ═══════════════════════════════════════════════════════
func (u *ProductUsecaseImpl) GetByProductID(ctx context.Context, id uint) (*entity.Product, error) {
	u.Log.WithField("id", id).Debug("Attempting to get product by ID")

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("id", id).Warn("Get product failed: product not found")
			return nil, apperror.ErrProductNotFound
		}

		u.Log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Get product failed: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	if !product.IsActive {
		u.Log.WithField("id", id).Warn("Get product failed: product not active")
		return nil, apperror.ErrProductNotFound
	}
	u.Log.WithField("id", id).Debug("Product retrieved")
	return product, nil
}

func (u *ProductUsecaseImpl) GetProductsByIDs(ctx context.Context, ids []uint) ([]entity.Product, error) {
	u.Log.WithField("ids", ids).Debug("Attempting to get products by IDs")
	products, err := u.ProductRepository.FindByIDs(ctx, ids)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"ids":   ids,
			"error": err,
		}).Error("Get products failed: unexpected DB error")
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
		u.Log.WithField("ids", ids).Warn("Get products failed: products not found")
		return nil, apperror.ErrProductNotFound
	}
	u.Log.WithField("ids", ids).Debug("Products retrieved")
	return result, nil
}

func (u *ProductUsecaseImpl) DecreaseStock(ctx context.Context, productID uint, quantity int) error {
	u.Log.WithFields(logrus.Fields{
		"product_id": productID,
		"quantity":   quantity,
	}).Debug("Attempting to decrease stock")
	err := u.ProductRepository.DecreaseStock(ctx, productID, quantity)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("product_id", productID).Warn("Decrease stock failed: product not found")
			return apperror.ErrProductNotFound
		}
		if errors.Is(err, apperror.ErrInsufficientStock) {
			u.Log.WithFields(logrus.Fields{
				"product_id": productID,
				"quantity":   quantity,
			}).Warn("Decrease stock failed: insufficient stock")
			return err
		}
		u.Log.WithFields(logrus.Fields{
			"product_id": productID,
			"error":      err,
		}).Error("Decrease stock failed: unexpected DB error")
		return apperror.ErrInternalServer
	}
	u.Log.WithField("product_id", productID).Debug("Stock decreased successfully")
	return nil
}
