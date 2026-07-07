package carthttp

import (
	"errors"
	"net/http"

	"github.com/Mpayy/e-commerce/internal/cart/dto"
	cartusecase "github.com/Mpayy/e-commerce/internal/cart/usecase"
	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type CartHandlerImpl struct {
	CartUsecase cartusecase.CartUsecase
}

func NewCartHandler(cartUsecase cartusecase.CartUsecase) CartHandler {
	return &CartHandlerImpl{CartUsecase: cartUsecase}
}

func (h *CartHandlerImpl) AddItem(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	var request dto.CartItem
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), err.Error())
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
