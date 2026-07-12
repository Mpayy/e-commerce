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

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Creates a product category and auto-generates its slug from the name. Requires admin role. Returns 409 if a category with the same name already exists.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CategoryRequest true "Category payload"
// @Success      201 {object} response.SuccessResponse{data=dto.CategoryResponse}
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      403 {object} response.ErrorResponse "Forbidden — admin role required"
// @Failure      409 {object} response.ErrorResponse "Category name already exists"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /admin/categories [post]
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

// GetAllCategories godoc
// @Summary      List all categories
// @Description  Returns every product category. This endpoint is public and does not require authentication.
// @Tags         categories
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=[]dto.CategoryResponse}
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /categories [get]
func (h *CategoryHandlerImpl) GetAll(ctx *gin.Context) {
	categories, err := h.CategoryUsecase.GetAllCategories(ctx.Request.Context())
	if err != nil {
		response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "Categories retrieved successfully", categories)
}
