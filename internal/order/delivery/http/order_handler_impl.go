package orderhttp

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/internal/order/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandlerImpl struct {
	OrderUsecase usecase.OrderUsecase
	Log          *logrus.Logger
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase, log *logrus.Logger) OrderHandler {
	return &OrderHandlerImpl{OrderUsecase: orderUsecase, Log: log}
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

	response.ResponseSuccess(ctx, http.StatusCreated, "Checkout successful", checkOutResponse)
}

func (h *OrderHandlerImpl) GetHistory(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	orders, err := h.OrderUsecase.GetOrderHistory(ctx.Request.Context(), auth.UserID)
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Order history retrieved successfully", orders)
}

func (h *OrderHandlerImpl) GetDetail(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	orderIDStr := ctx.Param("order_id")
	if orderIDStr == "" {
		response.ResponseError(ctx, http.StatusBadRequest, "Order ID is required", nil)
		return
	}
	
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.Log.WithField("error", err).Warn("Invalid order ID")
		response.ResponseError(ctx, http.StatusBadRequest, "Invalid order ID", nil)
		return
	}

	order, err := h.OrderUsecase.GetOrderDetail(ctx.Request.Context(), auth.UserID, uint(orderID))
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrOrderNotFound):
			response.ResponseError(ctx, http.StatusNotFound, apperror.ErrOrderNotFound.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Order detail retrieved successfully", order)
}
