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
		switch {
		case errors.Is(err, apperror.ErrCategoryNotFound):
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Product fetched successfully", products)
}