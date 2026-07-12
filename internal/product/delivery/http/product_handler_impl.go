package producthttp

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/internal/product/dto"
	productusecase "github.com/Mpayy/e-commerce/internal/product/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type ProductHandlerImpl struct {
	ProductUsecase productusecase.ProductUsecase
	Validator      *validator.Validate
	Log            *logrus.Logger
}

func NewProductHandler(productUsecase productusecase.ProductUsecase, validator *validator.Validate, log *logrus.Logger) ProductHandler {
	return &ProductHandlerImpl{ProductUsecase: productUsecase, Validator: validator, Log: log}
}

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Creates a product under an existing category. Slug is auto-generated from the name; SKU is auto-generated if left blank, otherwise the provided SKU is sanitized and used as-is. Requires admin role.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ProductCreateRequest true "Product payload"
// @Success      201 {object} response.SuccessResponse{data=dto.ProductResponse}
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Forbidden — admin role required"
// @Failure      404 {object} response.ErrorResponse "Category not found"
// @Failure      409 {object} response.ErrorResponse "Conflict (duplicate slug or SKU)"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /admin/products [post]
func (h *ProductHandlerImpl) Create(ctx *gin.Context) {
	var request dto.ProductCreateRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind json")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during create product")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	product, err := h.ProductUsecase.CreateProduct(ctx.Request.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrCategoryNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		case errors.Is(err, apperror.ErrDuplicatedProduct):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		case errors.Is(err, apperror.ErrDuplicatedProductSku):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusCreated, "Product created successfully", product)
}

// UpdateProduct godoc
// @Summary      Update an existing product
// @Description  Updates all fields of a product, including its active status. The category is re-validated only if category_id is changed. Requires admin role.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        product_id path int true "Product ID"
// @Param        request body dto.ProductUpdateRequest true "Product payload"
// @Success      200 {object} response.SuccessResponse{data=dto.ProductResponse}
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Forbidden — admin role required"
// @Failure      404 {object} response.ErrorResponse "Not Found (Product or Category not found)"
// @Failure      409 {object} response.ErrorResponse "Conflict (duplicate slug or SKU)"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /admin/products/{product_id} [put]
func (h *ProductHandlerImpl) Update(ctx *gin.Context) {
	var request dto.ProductUpdateRequest

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

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind json")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during update product")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	product, err := h.ProductUsecase.UpdateProduct(ctx.Request.Context(), uint(productID), &request)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		case errors.Is(err, apperror.ErrCategoryNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		case errors.Is(err, apperror.ErrDuplicatedProduct):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		case errors.Is(err, apperror.ErrDuplicatedProductSku):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Product updated successfully", product)
}

// DeleteProduct godoc
// @Summary      Soft-delete a product
// @Description  Deactivates a product by setting is_active to false instead of removing the row, so past order history referencing this product remains intact. Requires admin role.
// @Tags         products
// @Security     BearerAuth
// @Produce      json
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse "Product deleted successfully"
// @Failure      400 {object} response.ErrorResponse "Invalid product ID"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Forbidden — admin role required"
// @Failure      404 {object} response.ErrorResponse "Product not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /admin/products/{product_id} [delete]
func (h *ProductHandlerImpl) Delete(ctx *gin.Context) {
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

	err = h.ProductUsecase.DeleteProduct(ctx.Request.Context(), uint(productID))
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Product deleted successfully", nil)
}

// AdjustStock godoc
// @Summary      Set a product's stock quantity
// @Description  Overwrites the product's stock to the given absolute value (not a delta), typically used to reconcile stock after a physical count. Requires admin role.
// @Tags         products
// @Security     BearerAuth
// @Produce      json
// @Param        product_id path int true "Product ID"
// @Param        request body dto.ProductStockAdjustmentRequest true "Stock adjustment payload"
// @Success      200 {object} response.SuccessResponse "Product stock adjusted successfully"
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Forbidden — admin role required"
// @Failure      404 {object} response.ErrorResponse "Product not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /admin/products/{product_id}/adjust-stock [patch]
func (h *ProductHandlerImpl) AdjustStock(ctx *gin.Context) {
	var request dto.ProductStockAdjustmentRequest

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

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind json")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during adjust stock")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	err = h.ProductUsecase.AdjustStock(ctx.Request.Context(), uint(productID), *request.Stock)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Stock adjusted successfully", nil)
}

// GetByID godoc
// @Summary      Get product detail
// @Description  Returns a single product by ID. Inactive or non-existent products both return 404, so publicly disabled products are indistinguishable from products that were never created. This endpoint is public.
// @Tags         products
// @Produce      json
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse{data=dto.ProductResponse}
// @Failure      400 {object} response.ErrorResponse "Invalid product ID"
// @Failure      404 {object} response.ErrorResponse "Product not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /products/{product_id} [get]
func (h *ProductHandlerImpl) GetByID(ctx *gin.Context) {
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

	product, err := h.ProductUsecase.GetProductDetail(ctx.Request.Context(), uint(productID))
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrProductNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Product fetched successfully", product)
}

// Search godoc
// @Summary      Search and list products
// @Description  Returns a paginated, publicly accessible list of active products, optionally filtered by name (partial match) and category_id. A category_id that matches no products returns an empty list, not a 404.
// @Tags         products
// @Produce      json
// @Param        search      query string false "Search by product name"
// @Param        category_id query int    false "Filter by category ID"
// @Param        page        query int    false "Page number" default(1)
// @Param        limit       query int    false "Items per page" default(10)
// @Success      200 {object} response.SuccessResponse{data=dto.ProductSearchResponse}
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /products [get]
func (h *ProductHandlerImpl) Search(ctx *gin.Context) {
	var request dto.ProductSearchRequest

	if err := ctx.ShouldBindQuery(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind query")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during search product")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	if request.Page == 0 {
		request.Page = 1
	}
	if request.Limit == 0 {
		request.Limit = 10
	}

	products, err := h.ProductUsecase.SearchProducts(ctx.Request.Context(), &request)
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Product fetched successfully", products)
}
