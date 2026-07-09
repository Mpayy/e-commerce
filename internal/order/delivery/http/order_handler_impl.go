package orderhttp

import (
	"errors"
	"net/http"

	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/internal/order/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type OrderHandlerImpl struct {
	OrderUsecase usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase) OrderHandler {
	return &OrderHandlerImpl{OrderUsecase: orderUsecase}
}

func (h *OrderHandlerImpl) Checkout(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	checkOutResponse, err := h.OrderUsecase.Checkout(ctx.Request.Context(), auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrCartEmpty):
			response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrCartEmpty.Error(), nil)
			return
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, apperror.ErrProductNotFound.Error(), nil)
			return
		case errors.Is(err, apperror.ErrInsufficientStock):
			response.ResponseError(ctx, http.StatusUnprocessableEntity, apperror.ErrInsufficientStock.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Checkout successful", checkOutResponse)
}
