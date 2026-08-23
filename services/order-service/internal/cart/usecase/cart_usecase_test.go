package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repoMock "github.com/Mpayy/e-commerce/services/order-service/internal/cart/mocks"
	"github.com/Mpayy/e-commerce/services/order-service/internal/product/entity"
	productMock "github.com/Mpayy/e-commerce/services/order-service/internal/product/mocks"
)

func newTestLogger() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupCartUsecase(t *testing.T) (CartUsecase, *repoMock.MockCartRedisRepository, *productMock.MockProductService) {
	cartRepository := repoMock.NewMockCartRedisRepository(t)
	productService := productMock.NewMockProductService(t)
	log := newTestLogger()

	t.Cleanup(func() {
		cartRepository.AssertExpectations(t)
		productService.AssertExpectations(t)
	})

	cartUsecase := NewCartUsecase(cartRepository, productService, log)
	return cartUsecase, cartRepository, productService
}

func setupCartService(t *testing.T) (CartService, *repoMock.MockCartRedisRepository) {
	cartRepository := repoMock.NewMockCartRedisRepository(t)
	log := newTestLogger()

	t.Cleanup(func() {
		cartRepository.AssertExpectations(t)
	})

	cartService := NewCartUsecase(cartRepository, nil, log)
	return cartService, cartRepository
}

func TestCartUsecaseImpl_AddToCart(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	productID := uint(101)
	quantity := 2
	product := entity.Product{
		ID: productID,
	}

	t.Run("success_add_to_cart", func(t *testing.T) {
		cartUsecase, cartRepository, productService := setupCartUsecase(t)

		productService.EXPECT().
			GetByProductID(mock.Anything, productID).
			Return(&product, nil)

		cartRepository.EXPECT().
			AddItem(mock.Anything, userID, productID, quantity).
			Return(nil)

		err := cartUsecase.AddToCart(ctx, userID, productID, quantity)

		assert.NoError(t, err)
	})

	t.Run("error_invalid_quantity", func(t *testing.T) {
		cartUsecase, _, _ := setupCartUsecase(t)

		err := cartUsecase.AddToCart(ctx, userID, productID, 0)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperror.ErrInvalidQuantity)
	})

	t.Run("error_product_service_get_by_id_failed", func(t *testing.T) {
		cartUsecase, _, productService := setupCartUsecase(t)

		expectedErr := apperror.ErrProductNotFound
		productService.EXPECT().
			GetByProductID(mock.Anything, productID).
			Return(nil, expectedErr)

		// AddItem di Redis tidak boleh dipanggil jika product service error
		err := cartUsecase.AddToCart(ctx, userID, productID, quantity)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("error_cart_repository_add_item_failed", func(t *testing.T) {
		cartUsecase, cartRepository, productService := setupCartUsecase(t)

		redisErr := errors.New("redis connection timeout")

		productService.EXPECT().
			GetByProductID(mock.Anything, productID).
			Return(&product, nil)

		cartRepository.EXPECT().
			AddItem(mock.Anything, userID, productID, quantity).
			Return(redisErr)

		err := cartUsecase.AddToCart(ctx, userID, productID, quantity)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add item to cart")
		assert.ErrorIs(t, err, redisErr)
	})
}

func TestCartUsecaseImpl_UpdateCartItem(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	productID := uint(101)
	quantity := 5

	t.Run("success_update_cart_item", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		cartRepository.EXPECT().
			UpdateItem(mock.Anything, userID, productID, quantity).
			Return(nil)

		err := cartUsecase.UpdateCartItem(ctx, userID, productID, quantity)

		assert.NoError(t, err)
	})

	t.Run("quantity_zero_or_negative_redirects_to_remove_from_cart", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		// Saat quantity <= 0, UpdateCartItem akan memanggil RemoveFromCart
		cartRepository.EXPECT().
			RemoveItem(mock.Anything, userID, productID).
			Return(nil)

		err := cartUsecase.UpdateCartItem(ctx, userID, productID, 0)

		assert.NoError(t, err)
	})

	t.Run("error_item_not_found", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		cartRepository.EXPECT().
			UpdateItem(mock.Anything, userID, productID, quantity).
			Return(apperror.ErrRecordNotFound)

		err := cartUsecase.UpdateCartItem(ctx, userID, productID, quantity)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperror.ErrCartNotFound)
	})

	t.Run("error_repository_update_failed", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		redisErr := errors.New("redis connection error")

		cartRepository.EXPECT().
			UpdateItem(mock.Anything, userID, productID, quantity).
			Return(redisErr)

		err := cartUsecase.UpdateCartItem(ctx, userID, productID, quantity)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update item in cart")
		assert.ErrorIs(t, err, redisErr)
	})
}

func TestCartUsecaseImpl_RemoveFromCart(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	productID := uint(101)

	t.Run("success_remove_from_cart", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		cartRepository.EXPECT().
			RemoveItem(mock.Anything, userID, productID).
			Return(nil)

		err := cartUsecase.RemoveFromCart(ctx, userID, productID)

		assert.NoError(t, err)
	})

	t.Run("error_repository_remove_failed", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		redisErr := errors.New("redis error")

		cartRepository.EXPECT().
			RemoveItem(mock.Anything, userID, productID).
			Return(redisErr)

		err := cartUsecase.RemoveFromCart(ctx, userID, productID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove item from cart")
		assert.ErrorIs(t, err, redisErr)
	})
}

func TestCartUsecaseImpl_GetCartDetail(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	t.Run("error_get_cart_redis_failed", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		redisErr := errors.New("redis connection error")
		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(nil, redisErr)

		res, err := cartUsecase.GetCartDetail(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, redisErr)
		assert.Contains(t, err.Error(), "failed to get cart detail")
	})

	t.Run("success_empty_cart", func(t *testing.T) {
		cartUsecase, cartRepository, _ := setupCartUsecase(t)

		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(map[uint]int{}, nil)

		res, err := cartUsecase.GetCartDetail(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res.Items)
		assert.Empty(t, res.UnavailableItems)
		assert.Equal(t, float64(0), res.GrandTotal)
	})

	t.Run("error_product_not_found_returns_all_as_unavailable", func(t *testing.T) {
		cartUsecase, cartRepository, productService := setupCartUsecase(t)

		cartMap := map[uint]int{101: 2}
		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(cartMap, nil)

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(nil, apperror.ErrProductNotFound)

		res, err := cartUsecase.GetCartDetail(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Empty(t, res.Items)
		assert.Len(t, res.UnavailableItems, 1)
		assert.Equal(t, uint(101), res.UnavailableItems[0].ProductID)
		assert.Equal(t, "Produk sudah tidak tersedia atau dihapus", res.UnavailableItems[0].Message)
		assert.Equal(t, float64(0), res.GrandTotal)
	})

	t.Run("error_product_service_generic_error", func(t *testing.T) {
		cartUsecase, cartRepository, productService := setupCartUsecase(t)

		cartMap := map[uint]int{101: 2}
		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(cartMap, nil)

		grpcErr := errors.New("grpc connection failed")
		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(nil, grpcErr)

		res, err := cartUsecase.GetCartDetail(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, grpcErr)
	})

	t.Run("success_with_items_qty_zero_and_missing_products", func(t *testing.T) {
		cartUsecase, cartRepository, productService := setupCartUsecase(t)

		// cartMap berisi 3 skenario:
		// - 101: valid & ada di DB (qty 2)
		// - 102: qty <= 0 (uji percabangan if qty <= 0)
		// - 103: ada di Redis tapi hilang dari DB (uji percabangan unavailableItems)
		cartMap := map[uint]int{
			101: 2,
			102: 0,
			103: 1,
		}

		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(cartMap, nil)

		productsFromDB := []entity.Product{
			{ID: 101, Name: "Product 101", Price: 50000, Stock: 10},
			{ID: 102, Name: "Product 102", Price: 20000, Stock: 5},
		}

		productService.EXPECT().
			GetProductsByIDs(mock.Anything, mock.Anything).
			Return(productsFromDB, nil)

		res, err := cartUsecase.GetCartDetail(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Hanya Product 101 yang masuk Items (Product 102 terlewat karena qty 0)
		assert.Len(t, res.Items, 1)
		assert.Equal(t, uint(101), res.Items[0].ProductID)
		assert.Equal(t, "Product 101", res.Items[0].Name)
		assert.Equal(t, float64(50000), res.Items[0].Price)
		assert.Equal(t, 2, res.Items[0].Quantity)
		assert.Equal(t, float64(100000), res.Items[0].Subtotal)

		// Product 103 masuk ke UnavailableItems
		assert.Len(t, res.UnavailableItems, 1)
		assert.Equal(t, uint(103), res.UnavailableItems[0].ProductID)
		assert.Equal(t, "Product is not available", res.UnavailableItems[0].Message)
		assert.Equal(t, float64(100000), res.GrandTotal)
	})
}

func TestCartUsecaseImpl_GetRawCart(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	t.Run("success", func(t *testing.T) {
		cartService, cartRepository := setupCartService(t)

		expectedCart := map[uint]int{
			101: 2,
			102: 5,
		}

		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(expectedCart, nil)

		res, err := cartService.GetRawCart(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedCart, res)
	})

	t.Run("error_redis_get_cart_failed", func(t *testing.T) {
		cartService, cartRepository := setupCartService(t)

		redisErr := errors.New("redis connection failed")
		cartRepository.EXPECT().
			GetCart(mock.Anything, userID).
			Return(nil, redisErr)

		res, err := cartService.GetRawCart(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, redisErr)
		assert.Contains(t, err.Error(), "failed to get raw cart")
	})
}

func TestCartUsecaseImpl_ClearCart(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	t.Run("success", func(t *testing.T) {
		cartService, cartRepository := setupCartService(t)

		cartRepository.EXPECT().
			ClearCart(mock.Anything, userID).
			Return(nil)

		err := cartService.ClearCart(ctx, userID)

		assert.NoError(t, err)
	})

	t.Run("error_redis_clear_cart_failed", func(t *testing.T) {
		cartService, cartRepository := setupCartService(t)

		redisErr := errors.New("redis del command failed")
		cartRepository.EXPECT().
			ClearCart(mock.Anything, userID).
			Return(redisErr)

		err := cartService.ClearCart(ctx, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, redisErr)
		assert.Contains(t, err.Error(), "failed to clear cart")
	})
}