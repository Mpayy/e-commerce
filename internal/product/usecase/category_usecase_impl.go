package productusecase

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/internal/product/dto"
	"github.com/Mpayy/e-commerce/internal/product/entity"
	productrepository "github.com/Mpayy/e-commerce/internal/product/repository"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

type CategoryUsecaseImpl struct {
	CategoryRepo productrepository.CategoryRepository
	Log          *logrus.Logger
	Transaction  transaction.Transaction
}

func NewCategoryUsecase(categoryRepo productrepository.CategoryRepository, log *logrus.Logger, transaction transaction.Transaction) CategoryUsecase {
	return &CategoryUsecaseImpl{CategoryRepo: categoryRepo, Log: log, Transaction: transaction}
}

func (u *CategoryUsecaseImpl) CreateCategory(ctx context.Context, request *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	u.Log.WithField("name", request.Name).Debug("Attempting to create category")

	category := &entity.Category{
		Name: request.Name,
		Slug: slug.Make(request.Name),
	}

	err := u.Transaction.WithTransaction(ctx, func(ctx context.Context) error {
		errCreate := u.CategoryRepo.Create(ctx, category)
		if errCreate != nil {
			if errors.Is(errCreate, apperror.ErrDuplicatedKey) {
				u.Log.WithField("name", request.Name).
					Warn("Create category failed: duplicate name")
				return apperror.ErrDuplicatedCategory
			}
			u.Log.WithFields(logrus.Fields{
				"name":  request.Name,
				"error": errCreate,
			}).Error("Create category failed: unexpected DB error")
			return apperror.ErrInternalServer
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	response := &dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}

	u.Log.WithField("name", request.Name).Info("Category created successfully")
	return response, nil
}
