package carthttp

import (
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/services/order-service/internal/cart/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CartHandlerImpl struct {
	cartUsecase usecase.CartUsecase
	cartService usecase.CartService
	validator   *validator.Validate
}

func NewCartHandler(cartUsecase usecase.CartUsecase, cartService usecase.CartService, validator *validator.Validate) CartHandler {
	return &CartHandlerImpl{cartUsecase: cartUsecase, cartService: cartService, validator: validator}
}

// AddItemCart godoc
// @Summary      Add an item to the cart
// @Description  Adds a product and quantity to the authenticated user's Redis-backed cart. If the product is already in the cart, the quantity is incremented rather than overwritten.
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CartItemCreateRequest true "Cart payload"
// @Success      200 {object} response.SuccessResponse "Item added to cart successfully"
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Validation error"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "Product not found"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /cart [post]
func (h *CartHandlerImpl) AddItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	var request dto.CartItemCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	err := h.cartUsecase.AddToCart(ctx.Request.Context(), auth.ID, request.ProductID, request.Quantity)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, nil)
}

// UpdateItemCart godoc
// @Summary      Update a cart item's quantity
// @Description  Overwrites the quantity of a product already in the cart with the given value. Returns 404 if the product was never added to this cart. Sending a quantity of 0 removes the item instead.
// @Tags         carts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        product_id path int true "Product ID"
// @Param        request body dto.CartItemUpdateRequest true "Cart payload"
// @Success      200 {object} response.SuccessResponse "Item updated in cart successfully"
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Validation error"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "Cart item not found"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /cart/{product_id} [patch]
func (h *CartHandlerImpl) UpdateItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	productIDParam := ctx.Param("product_id")
	if productIDParam == "" {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	var request dto.CartItemUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	err = h.cartUsecase.UpdateCartItem(ctx.Request.Context(), auth.ID, uint(productID), *request.Quantity)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.ResponseSuccess(ctx, http.StatusOK, nil)
}

// RemoveItemCart godoc
// @Summary      Remove a single item from the cart
// @Description  Removes one product from the authenticated user's cart by product ID. Removing a product that isn't in the cart is treated as a no-op, not an error.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse "Item removed from cart successfully"
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Invalid product ID"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /cart/{product_id} [delete]
func (h *CartHandlerImpl) RemoveItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	productIDParam := ctx.Param("product_id")
	if productIDParam == "" {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	err = h.cartUsecase.RemoveFromCart(ctx.Request.Context(), auth.ID, uint(productID))
	if err != nil {
		response.HandleError(ctx, apperror.ErrInternalServer)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, nil)
}

// GetCart godoc
// @Summary      Get the authenticated user's cart
// @Description  Returns cart items enriched with live product name, price, and stock via a single bulk lookup, along with the computed grand total. Products removed from the catalog since being added are silently excluded.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse{data=dto.CartDetailResponse} "cart detail retrieved successfully"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /cart [get]
func (h *CartHandlerImpl) GetCart(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	cartDetail, err := h.cartUsecase.GetCartDetail(ctx.Request.Context(), auth.ID)
	if err != nil {
		response.HandleError(ctx, apperror.ErrInternalServer)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, cartDetail)
}

// ClearCart godoc
// @Summary      Empty the cart
// @Description  Removes all items from the authenticated user's cart in a single operation.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse "cart cleared successfully"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /cart [delete]
func (h *CartHandlerImpl) ClearCart(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	err := h.cartService.ClearCart(ctx.Request.Context(), auth.ID)
	if err != nil {
		response.HandleError(ctx, apperror.ErrInternalServer)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, nil)
}
