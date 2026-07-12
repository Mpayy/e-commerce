package carthttp

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/internal/cart/dto"
	cartusecase "github.com/Mpayy/e-commerce/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CartHandlerImpl struct {
	CartUsecase cartusecase.CartUsecase
	CartService cartusecase.CartService
	Validator   *validator.Validate
	Log         *logrus.Logger
}

func NewCartHandler(cartUsecase cartusecase.CartUsecase, cartService cartusecase.CartService, validator *validator.Validate, log *logrus.Logger) CartHandler {
	return &CartHandlerImpl{CartUsecase: cartUsecase, CartService: cartService, Validator: validator, Log: log}
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
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "Product not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /cart [post]
func (h *CartHandlerImpl) AddItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	var request dto.CartItemCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind JSON")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during add item")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	err := h.CartUsecase.AddToCart(ctx.Request.Context(), auth.UserID, request.ProductID, request.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidQuantity):
			response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrInvalidQuantity.Error(), nil)
			return
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, apperror.ErrProductNotFound.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "item added to cart successfully", nil)
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
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "Cart item not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /cart/{product_id} [patch]
func (h *CartHandlerImpl) UpdateItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	productIDParam := ctx.Param("product_id")
	if productIDParam == "" {
		response.ResponseError(ctx, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		h.Log.WithField("error", err).Warn("Invalid product ID")
		response.ResponseError(ctx, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	var request dto.CartItemUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind JSON")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during update item")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	err = h.CartUsecase.UpdateCartItem(ctx.Request.Context(), auth.UserID, uint(productID), *request.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrCartNotFound):
			response.ResponseError(ctx, http.StatusNotFound, apperror.ErrCartNotFound.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}
	}
	response.ResponseSuccess(ctx, http.StatusOK, "item updated in cart successfully", nil)
}

// RemoveItemCart godoc
// @Summary      Remove a single item from the cart
// @Description  Removes one product from the authenticated user's cart by product ID. Removing a product that isn't in the cart is treated as a no-op, not an error.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse "Item removed from cart successfully"
// @Failure      400 {object} response.ErrorResponse "Invalid product ID"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /cart/{product_id} [delete]
func (h *CartHandlerImpl) RemoveItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	productIDParam := ctx.Param("product_id")
	if productIDParam == "" {
		response.ResponseError(ctx, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		h.Log.WithField("error", err).Warn("Invalid product ID")
		response.ResponseError(ctx, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	err = h.CartUsecase.RemoveFromCart(ctx.Request.Context(), auth.UserID, uint(productID))
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "item removed from cart successfully", nil)
}

// GetCart godoc
// @Summary      Get the authenticated user's cart
// @Description  Returns cart items enriched with live product name, price, and stock via a single bulk lookup, along with the computed grand total. Products removed from the catalog since being added are silently excluded.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse{data=dto.CartDetailResponse} "cart detail retrieved successfully"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /cart [get]
func (h *CartHandlerImpl) GetCart(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	cartDetail, err := h.CartUsecase.GetCartDetail(ctx.Request.Context(), auth.UserID)
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "cart detail retrieved successfully", cartDetail)
}

// ClearCart godoc
// @Summary      Empty the cart
// @Description  Removes all items from the authenticated user's cart in a single operation.
// @Tags         carts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse "cart cleared successfully"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /cart [delete]
func (h *CartHandlerImpl) ClearCart(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	err := h.CartService.ClearCart(ctx.Request.Context(), auth.UserID)
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "cart cleared successfully", nil)
}
