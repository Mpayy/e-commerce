package http

import (
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	_ "github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderHandlerImpl struct {
	orderUsecase usecase.OrderUsecase
	validator    *validator.Validate
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase, validator *validator.Validate) OrderHandler {
	return &OrderHandlerImpl{orderUsecase: orderUsecase, validator: validator}
}

// CheckoutOrder godoc
// @Summary      Checkout the current cart into an order
// @Description  Converts the cart into a permanent order: validates and decrements stock for each item under a row lock inside a single database transaction, snapshots product name and price at time of purchase, then clears the cart only after the transaction commits successfully. If any item fails validation, the entire order is rolled back and the cart is left untouched so the user can retry.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} response.SuccessResponse{data=dto.OrderResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "CART_EMPTY"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND"
// @Failure      422 {object} response.ErrorResponse{error=apperror.AppError} "INSUFFICIENT_STOCK"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /orders [post]
func (h *OrderHandlerImpl) Checkout(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	checkOutResponse, err := h.orderUsecase.Checkout(ctx.Request.Context(), auth.ID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusCreated, checkOutResponse)
}

// GetOrderHistory godoc
// @Summary      List the authenticated user's past orders
// @Description  Returns a paginated list of every order placed by the authenticated user along with their line items, using the price and product name captured at checkout time rather than current catalog data.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} response.SuccessResponse{data=dto.OrderHistoryResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / VALIDATION_FAILED"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /orders [get]
func (h *OrderHandlerImpl) GetHistory(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	var request dto.OrderFilter
	if err := ctx.ShouldBindQuery(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	orders, err := h.orderUsecase.GetOrderHistory(ctx.Request.Context(), auth.ID, &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, orders)
}

// GetOrderDetail godoc
// @Summary      Get a single order's detail
// @Description  Returns the full detail of one order, including its line items. Returns 404 both when the order doesn't exist and when it belongs to a different user, so ownership is never leaked through the error response.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        order_id path int true "Order ID"
// @Success      200 {object} response.SuccessResponse{data=dto.OrderResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "ORDER_NOT_FOUND"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /orders/{order_id} [get]
func (h *OrderHandlerImpl) GetDetail(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	orderIDStr := ctx.Param("order_id")
	if orderIDStr == "" {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	order, err := h.orderUsecase.GetOrderDetail(ctx.Request.Context(), auth.ID, uint(orderID))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, order)
}
