package http

import (
	"net/http"
	"strconv"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/dto"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/entity"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProductHandlerImpl struct {
	productUsecase usecase.ProductUsecase
	validator      *validator.Validate
}

func NewProductHandler(productUsecase usecase.ProductUsecase, validator *validator.Validate) ProductHandler {
	return &ProductHandlerImpl{productUsecase: productUsecase, validator: validator}
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
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / VALIDATION_FAILED"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "FORBIDDEN"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "CATEGORY_NOT_FOUND"
// @Failure      409 {object} response.ErrorResponse{error=apperror.AppError} "DUPLICATED_PRODUCT / DUPLICATED_PRODUCT_SKU"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /admin/products [post]
func (h *ProductHandlerImpl) Create(ctx *gin.Context) {
	var request dto.ProductCreateRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	product, err := h.productUsecase.CreateProduct(ctx.Request.Context(), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusCreated, product)
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
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / VALIDATION_FAILED"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "FORBIDDEN"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND / CATEGORY_NOT_FOUND"
// @Failure      409 {object} response.ErrorResponse{error=apperror.AppError} "DUPLICATED_PRODUCT / DUPLICATED_PRODUCT_SKU"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /admin/products/{product_id} [put]
func (h *ProductHandlerImpl) Update(ctx *gin.Context) {
	var request dto.ProductUpdateRequest

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

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	product, err := h.productUsecase.UpdateProduct(ctx.Request.Context(), uint(productID), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary      Soft-delete a product
// @Description  Deactivates a product by setting is_active to false instead of removing the row, so past order history referencing this product remains intact. Requires admin role.
// @Tags         products
// @Security     BearerAuth
// @Produce      json
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse "Product deleted successfully"
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "FORBIDDEN"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /admin/products/{product_id} [delete]
func (h *ProductHandlerImpl) Delete(ctx *gin.Context) {
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

	err = h.productUsecase.DeleteProduct(ctx.Request.Context(), uint(productID))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, nil)
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
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / VALIDATION_FAILED"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "FORBIDDEN"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /admin/products/{product_id}/adjust-stock [patch]
func (h *ProductHandlerImpl) AdjustStock(ctx *gin.Context) {
	var request dto.ProductStockAdjustmentRequest

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

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	err = h.productUsecase.AdjustStock(ctx.Request.Context(), uint(productID), *request.Stock)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, nil)
}

// GetByID godoc
// @Summary      Get product detail
// @Description  Returns a single product by ID. Inactive or non-existent products both return 404, so publicly disabled products are indistinguishable from products that were never created. image_url is null if no image has been uploaded yet. This endpoint is public.
// @Tags         products
// @Produce      json
// @Param        product_id path int true "Product ID"
// @Success      200 {object} response.SuccessResponse{data=dto.ProductResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /products/{product_id} [get]
func (h *ProductHandlerImpl) GetByID(ctx *gin.Context) {
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

	product, err := h.productUsecase.GetProductDetail(ctx.Request.Context(), uint(productID))
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, product)
}

// Search godoc
// @Summary      Search and list products
// @Description  Returns a paginated, publicly accessible list of active products, optionally filtered by name (partial match) and category_id. A category_id that matches no products returns an empty list, not a 404. image_url is null for products without an uploaded image.
// @Tags         products
// @Produce      json
// @Param        search      query string false "Search by product name"
// @Param        category_id query int    false "Filter by category ID"
// @Param        page        query int    false "Page number" default(1)
// @Param        limit       query int    false "Items per page" default(10)
// @Success      200 {object} response.SuccessResponse{data=dto.ProductSearchResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / VALIDATION_FAILED"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /products [get]
func (h *ProductHandlerImpl) Search(ctx *gin.Context) {
	var request dto.ProductSearchRequest

	if err := ctx.ShouldBindQuery(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	products, err := h.productUsecase.SearchProducts(ctx.Request.Context(), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, products)
}

// UploadProductImage godoc
// @Summary      Upload or replace a product's image (Admin)
// @Description  Uploads a single image for a product, replacing any existing one. The file is validated by inspecting its actual byte content (not the filename extension or Content-Type header) — only JPEG, PNG, and WEBP are accepted, max 2MB. The new file is written to disk before the database record is updated, and the previous image (if any) is deleted only after the update succeeds — this ordering ensures a failure at any step never leaves the database pointing to a file that doesn't exist. Requires admin role.
// @Tags         admin-products
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        product_id path int true "Product ID"
// @Param        image formData file true "Image file (JPEG, PNG, or WEBP, max 2MB)"
// @Success      200 {object} response.SuccessResponse{data=dto.ProductResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "BAD_REQUEST / INVALID_FILE_TYPE / FILE_TOO_LARGE"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "UNAUTHORIZED"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "FORBIDDEN"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "PRODUCT_NOT_FOUND"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "INTERNAL_SERVER_ERROR"
// @Router       /admin/products/{product_id}/image [post]
func (h *ProductHandlerImpl) UploadProductImage(ctx *gin.Context) {
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

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, entity.MaxImageSize)
	if err := ctx.Request.ParseMultipartForm(entity.MaxImageSize); err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			response.HandleError(ctx, apperror.ErrFileTooLarge)
			return
		}
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	result, err := h.productUsecase.UploadProductImage(ctx.Request.Context(), uint(productID), fileHeader)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, result)
}
