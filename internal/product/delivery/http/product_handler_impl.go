package producthttp

import (
	"errors"
	"net/http"

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
	var request dto.ProductRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind json")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		h.Log.WithField("error", err).Warn("Validation error during create product")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), nil)
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
