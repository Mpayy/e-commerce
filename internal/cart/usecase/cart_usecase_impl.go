package cartusecase

import (
	"context"

	cartrepository "github.com/Mpayy/e-commerce/internal/cart/repository"
	productusecase "github.com/Mpayy/e-commerce/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/sirupsen/logrus"
)

type CartUsecaseImpl struct {
	CartRedisRepository cartrepository.CartRedisRepository
	ProductService      productusecase.ProductService
	Log                 *logrus.Logger
}

func NewCartUsecase(cartRedisRepository cartrepository.CartRedisRepository, productService productusecase.ProductService, log *logrus.Logger) CartUsecase {
	return &CartUsecaseImpl{CartRedisRepository: cartRedisRepository, ProductService: productService, Log: log}
}

func (u *CartUsecaseImpl) AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error {
	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
		"quantity":   quantity,
	}).Info("Adding item to cart")
	if quantity <= 0 {
		return apperror.ErrInvalidQuantity
	}

	product, err := u.ProductService.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	// TODO: ini di comment karena stock akan dicek saat checkout / soft warning
	// if product.Stock < quantity {
	// 	return apperror.ErrInsufficientStock
	// }

	err = u.CartRedisRepository.AddItem(ctx, userID, product.ID, quantity)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id":    userID,
			"product_id": productID,
			"error":      err,
		}).Error("Failed to add item to cart")
		return apperror.ErrInternalServer
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
	}).Debug("Item added to cart successfully")

	return nil
}

func (u *CartUsecaseImpl) UpdateCartItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
		"quantity":   quantity,
	}).Info("Updating item in cart")

	if quantity <= 0 {
		return u.RemoveFromCart(ctx, userID, productID)
	}

	err := u.CartRedisRepository.UpdateItem(ctx, userID, productID, quantity)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id":    userID,
			"product_id": productID,
			"error":      err,
		}).Error("Failed to update item in cart")
		return apperror.ErrInternalServer
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
	}).Debug("Item updated in cart successfully")

	return nil
}

func (u *CartUsecaseImpl) RemoveFromCart(ctx context.Context, userID uint, productID uint) error {
	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
	}).Info("Removing item from cart")

	err := u.CartRedisRepository.RemoveItem(ctx, userID, productID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id":    userID,
			"product_id": productID,
			"error":      err,
		}).Error("Failed to remove item from cart")
		return apperror.ErrInternalServer
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":    userID,
		"product_id": productID,
	}).Debug("Item removed from cart successfully")

	return nil
}
