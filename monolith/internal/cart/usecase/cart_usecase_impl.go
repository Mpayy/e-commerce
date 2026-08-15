package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mpayy/e-commerce/monolith/internal/cart/dto"
	"github.com/Mpayy/e-commerce/monolith/internal/cart/repository"
	"github.com/Mpayy/e-commerce/monolith/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
)

type CartUsecaseImpl struct {
	cartRedisRepository repository.CartRedisRepository
	productService      usecase.ProductService
	log                 *logger.Logger
}

func NewCartUsecase(cartRedisRepository repository.CartRedisRepository, productService usecase.ProductService, log *logger.Logger) *CartUsecaseImpl {
	return &CartUsecaseImpl{cartRedisRepository: cartRedisRepository, productService: productService, log: log}
}

func (u *CartUsecaseImpl) AddToCart(ctx context.Context, userID uint, productID uint, quantity int) error {
	log := u.log.WithFields(logger.Fields{
		"user_id":    userID,
		"product_id": productID,
		"quantity":   quantity,
	})
	log.Debug("Adding item to cart")

	if quantity <= 0 {
		return apperror.ErrInvalidQuantity
	}

	product, err := u.productService.GetByProductID(ctx, productID)
	if err != nil {
		return err
	}

	err = u.cartRedisRepository.AddItem(ctx, userID, product.ID, quantity)
	if err != nil {
		return fmt.Errorf("failed to add item to cart: %v", err)
	}

	log.Debug("Item added to cart successfully")
	return nil
}

func (u *CartUsecaseImpl) UpdateCartItem(ctx context.Context, userID uint, productID uint, quantity int) error {
	logger := u.log.WithFields(logger.Fields{
		"user_id":    userID,
		"product_id": productID,
		"quantity":   quantity,
	})
	logger.Debug("Updating item in cart")

	if quantity <= 0 {
		return u.RemoveFromCart(ctx, userID, productID)
	}

	err := u.cartRedisRepository.UpdateItem(ctx, userID, productID, quantity)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			logger.Warn("Cart item not found to update")
			return apperror.ErrCartNotFound
		}
		return fmt.Errorf("failed to update item in cart: %v", err)
	}

	logger.Debug("Item updated in cart successfully")
	return nil
}

func (u *CartUsecaseImpl) RemoveFromCart(ctx context.Context, userID uint, productID uint) error {
	logger := u.log.WithFields(logger.Fields{
		"user_id":    userID,
		"product_id": productID,
	})
	logger.Debug("Removing item from cart")

	err := u.cartRedisRepository.RemoveItem(ctx, userID, productID)
	if err != nil {
		return fmt.Errorf("failed to remove item from cart: %v", err)
	}

	logger.Debug("Item removed from cart successfully")
	return nil
}

func (u *CartUsecaseImpl) GetCartDetail(ctx context.Context, userID uint) (*dto.CartDetailResponse, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Getting cart detail")

	cartMap, err := u.cartRedisRepository.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart detail: %v", err)
	}

	var productIDs []uint
	for productID := range cartMap {
		productIDs = append(productIDs, productID)
	}

	if len(productIDs) == 0 {
		return &dto.CartDetailResponse{
			Items:            []dto.CartItemResponse{},
			UnavailableItems: []dto.CartUnavailableItemResp{},
			GrandTotal:       0,
		}, nil
	}

	products, err := u.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		if errors.Is(err, apperror.ErrProductNotFound) {
			var unavailableItems []dto.CartUnavailableItemResp
			for pID := range cartMap {
				unavailableItems = append(unavailableItems, dto.CartUnavailableItemResp{
					ProductID: pID,
					Message:   "Produk sudah tidak tersedia atau dihapus",
				})
			}
			return &dto.CartDetailResponse{
				Items:            []dto.CartItemResponse{},
				UnavailableItems: unavailableItems,
				GrandTotal:       0,
			}, nil
		}
		return nil, err
	}

	foundProductsInDB := make(map[uint]bool)
	var itemsResponse []dto.CartItemResponse
	var grandTotal float64
	for _, product := range products {
		qty := cartMap[product.ID]
		if qty <= 0 {
			continue
		}

		foundProductsInDB[product.ID] = true
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

	unavailableItems := []dto.CartUnavailableItemResp{}
	for redisProductID := range cartMap {
		if !foundProductsInDB[redisProductID] {
			unavailableItems = append(unavailableItems, dto.CartUnavailableItemResp{
				ProductID: redisProductID,
				Message:   "Product is not available",
			})
		}
	}

	log.WithFields(logger.Fields{
		"items":             len(itemsResponse),
		"unavailable_items": len(unavailableItems),
		"grand_total":       grandTotal,
	}).Debug("Cart detail retrieved successfully")

	return &dto.CartDetailResponse{
		Items:            itemsResponse,
		UnavailableItems: unavailableItems,
		GrandTotal:       grandTotal,
	}, nil
}

// ═══════════════════════════════════════════════════════
// Consumption By Other Services (contract.go)
// ═══════════════════════════════════════════════════════
func (u *CartUsecaseImpl) GetRawCart(ctx context.Context, userID uint) (map[uint]int, error) {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Getting raw cart")

	cart, err := u.cartRedisRepository.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw cart: %v", err)
	}

	log.Debug("Raw cart retrieved successfully")
	return cart, nil
}

func (u *CartUsecaseImpl) ClearCart(ctx context.Context, userID uint) error {
	log := u.log.WithFields(logger.Fields{
		"user_id": userID,
	})
	log.Debug("Clearing cart")

	err := u.cartRedisRepository.ClearCart(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %v", err)
	}

	log.Debug("Cart cleared successfully")
	return nil
}
