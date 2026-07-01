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

type CategoryHandlerImpl struct {
	CategoryUsecase productusecase.CategoryUsecase
	Validator       *validator.Validate
	Log             *logrus.Logger
}

func NewCategoryHandler(categoryUsecase productusecase.CategoryUsecase, validator *validator.Validate, log *logrus.Logger) CategoryHandler {
	return &CategoryHandlerImpl{
		CategoryUsecase: categoryUsecase,
		Validator:       validator,
		Log:             log,
	}
}

func (h *CategoryHandlerImpl) Create(ctx *gin.Context) {
	var request dto.CategoryRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.Log.WithField("error", err).Warn("Failed to bind JSON")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithField("error", err).Warn("Validation error during create category")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	category, err := h.CategoryUsecase.CreateCategory(ctx.Request.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrDuplicatedCategory):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusCreated, "Category created successfully", category)
}
