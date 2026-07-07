package cartusecase

import (
	"context"
	"errors"

	cartrepository "github.com/Mpayy/e-commerce/internal/cart/repository"
	productusecase "github.com/Mpayy/e-commerce/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
)

type CartUsecaseImpl struct {
	CartRedisRepository cartrepository.CartRedisRepository
	ProductService    productusecase.ProductService
}

func NewCartUsecase(cartRedisRepository cartrepository.CartRedisRepository, productService productusecase.ProductService) CartUsecase {
	return &CartUsecaseImpl{CartRedisRepository: cartRedisRepository, ProductService: productService}
}

func (u *CartUsecaseImpl) AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error {
	if quantity <= 0 {
		return apperror.ErrInvalidQuantity
	}

	product, err := u.ProductService.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, apperror.ErrProductNotFound) {
			return apperror.ErrProductNotFound
		}
		return err
	}

	// TODO: ini di comment karena stock akan dicek saat checkout / soft warning
	// if product.Stock < quantity {
	// 	return apperror.ErrInsufficientStock
	// }

	err = u.CartRedisRepository.AddItem(ctx, userID, product.ID, quantity)
	if err != nil {
		return apperror.ErrInternalServer
	}

	return nil
}
