package http

import (
	"net/http"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/dto"
	"github.com/Mpayy/e-commerce/services/product-service/internal/product/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CategoryHandlerImpl struct {
	categoryUsecase usecase.CategoryUsecase
	validator       *validator.Validate
}

func NewCategoryHandler(categoryUsecase usecase.CategoryUsecase, validator *validator.Validate) CategoryHandler {
	return &CategoryHandlerImpl{categoryUsecase: categoryUsecase, validator: validator}
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
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Validation error"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      403 {object} response.ErrorResponse{error=apperror.AppError} "Forbidden — admin role required"
// @Failure      409 {object} response.ErrorResponse{error=apperror.AppError} "Category name already exists"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /admin/categories [post]
func (h *CategoryHandlerImpl) Create(ctx *gin.Context) {
	var request dto.CategoryRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	category, err := h.categoryUsecase.CreateCategory(ctx.Request.Context(), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusCreated, category)
}

// GetAllCategories godoc
// @Summary      List all categories
// @Description  Returns every product category. This endpoint is public and does not require authentication.
// @Tags         categories
// @Produce      json
// @Success      200 {object} response.SuccessResponse{data=[]dto.CategoryResponse}
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /categories [get]
func (h *CategoryHandlerImpl) GetAll(ctx *gin.Context) {
	categories, err := h.categoryUsecase.GetAllCategories(ctx.Request.Context())
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, categories)
}
