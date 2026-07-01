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

func NewProductUsecase(productRepository productrepository.ProductRepository, categoryUsecase CategoryUsecase, log *logrus.Logger, transaction transaction.Transaction) ProductUsecase {
	return &ProductUsecaseImpl{
		ProductRepository: productRepository,
		CategoryUsecase:   categoryUsecase,
		Log:               log,
		Transaction:       transaction,
	}
}

func (u *ProductUsecaseImpl) CreateProduct(ctx context.Context, request *dto.ProductRequest) (*dto.ProductResponse, error) {
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

	return &dto.ProductResponse{
		ID:         product.ID,
		CategoryID: product.CategoryID,
		Name:       product.Name,
		Slug:       product.Slug,
		Price:      product.Price,
		Stock:      product.Stock,
		SKU:        product.SKU,
		IsActive:   product.IsActive,
	}, nil
}
