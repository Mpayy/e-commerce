package cartusecase

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/internal/cart/dto"
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

func (u *CartUsecaseImpl) GetCartDetail(ctx context.Context, userID uint) (*dto.CartDetailResponse, error) {
	u.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("Getting cart detail")

	cartMap, err := u.CartRedisRepository.GetCart(ctx, userID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get cart detail")
		return nil, apperror.ErrInternalServer
	}

	var productIDs []uint
	for productID := range cartMap {
		productIDs = append(productIDs, productID)
	}

	if len(productIDs) == 0 {
		return &dto.CartDetailResponse{
			Items:      []dto.CartItemResponse{},
			GrandTotal: 0,
		}, nil
	}

	products, err := u.ProductService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		if errors.Is(err, apperror.ErrProductNotFound) {
			return &dto.CartDetailResponse{
				Items:      []dto.CartItemResponse{},
				GrandTotal: 0,
			}, nil
		}
		return nil, err
	}

	var itemsResponse []dto.CartItemResponse
	var grandTotal float64
	for _, product := range products {
		qty := cartMap[product.ID]
		if qty <= 0 {
			continue
		}

		subtotal := product.Price * float64(qty)
		grandTotal += subtotal

		itemsResponse = append(itemsResponse, dto.CartItemResponse{
			ProductID:      product.ID,
			Name:           product.Name,
			Price:          product.Price,
			Quantity:       qty,
			Subtotal:       subtotal,
			StockAvailable: product.Stock,
		})
	}

	u.Log.WithFields(logrus.Fields{
		"user_id":     userID,
		"items":       len(itemsResponse),
		"grand_total": grandTotal,
	}).Debug("Cart detail retrieved successfully")

	return &dto.CartDetailResponse{
		Items:      itemsResponse,
		GrandTotal: grandTotal,
	}, nil
}

// ═══════════════════════════════════════════════════════
// Consumption By Other Services (contract.go)
// ═══════════════════════════════════════════════════════

func (u *CartUsecaseImpl) GetRawCart(ctx context.Context, userID uint) (map[uint]int, error) {
	u.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("Getting raw cart")

	cart, err := u.CartRedisRepository.GetCart(ctx, userID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get raw cart")
		return nil, apperror.ErrInternalServer
	}

	u.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Debug("Raw cart retrieved successfully")

	return cart, nil
}

func (u *CartUsecaseImpl) ClearCart(ctx context.Context, userID uint) error {
	u.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("Clearing cart")

	err := u.CartRedisRepository.ClearCart(ctx, userID)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to clear cart")
		return apperror.ErrInternalServer
	}

	u.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Debug("Cart cleared successfully")

	return nil
}
