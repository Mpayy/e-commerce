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
	Validator   *validator.Validate
	Log         *logrus.Logger
}

func NewCartHandler(cartUsecase cartusecase.CartUsecase, validator *validator.Validate, log *logrus.Logger) CartHandler {
	return &CartHandlerImpl{CartUsecase: cartUsecase, Validator: validator, Log: log}
}

func (h *CartHandlerImpl) AddItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	var request dto.CartItemCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind JSON")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), err.Error())
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

	err = h.CartUsecase.UpdateCartItem(ctx.Request.Context(), auth.UserID, uint(productID), request.Quantity)
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "item updated in cart successfully", nil)
}

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
