package orderhttp

import (
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/monolith/internal/order/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type OrderHandlerImpl struct {
	orderUsecase usecase.OrderUsecase
	log          *logger.Logger
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase, log *logger.Logger) OrderHandler {
	return &OrderHandlerImpl{orderUsecase: orderUsecase, log: log}
}

// CheckoutOrder godoc
// @Summary      Checkout the current cart into an order
// @Description  Converts the cart into a permanent order: validates and decrements stock for each item under a row lock inside a single database transaction, snapshots product name and price at time of purchase, then clears the cart only after the transaction commits successfully. If any item fails validation, the entire order is rolled back and the cart is left untouched so the user can retry.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Success      201 {object} response.SuccessResponse{data=dto.OrderResponse}
// @Failure      400 {object} response.ErrorResponse "Cart is empty"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "Product not found"
// @Failure      422 {object} response.ErrorResponse "Insufficient stock"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
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
// @Description  Returns every order placed by the authenticated user along with their line items, using the price and product name captured at checkout time rather than current catalog data.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse{data=dto.OrderHistoryResponse}
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /orders [get]
func (h *OrderHandlerImpl) GetHistory(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	orders, err := h.orderUsecase.GetOrderHistory(ctx.Request.Context(), auth.ID)
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
// @Failure      400 {object} response.ErrorResponse "Invalid order ID"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "Order not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
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
