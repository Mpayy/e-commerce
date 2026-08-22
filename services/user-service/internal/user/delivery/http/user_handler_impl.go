package http

import (
	"net/http"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/middleware"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/dto"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandlerImpl struct {
	userUsecase usecase.UserUsecase
	validator   *validator.Validate
}

func NewUserHandler(userUsecase usecase.UserUsecase, validator *validator.Validate) UserHandler {
	return &UserHandlerImpl{
		userUsecase: userUsecase,
		validator:   validator,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with a bcrypt-hashed password. Email must be unique; returns 409 if already registered.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.UserRegisterRequest true "User payload"
// @Success      201 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Validation error"
// @Failure      409 {object} response.ErrorResponse{error=apperror.AppError} "Email already exists"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /register [post]
func (h *UserHandlerImpl) Register(ctx *gin.Context) {
	var request dto.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	user, err := h.userUsecase.Register(ctx.Request.Context(), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusCreated, user)
}

// Login godoc
// @Summary      Login and obtain a JWT token
// @Description  Authenticates a user by email and password, then issues a JWT stored in Redis for session tracking. Returns a generic error for both wrong password and unknown email to avoid leaking which emails are registered.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.UserLoginRequest true "User payload"
// @Success      200 {object} response.SuccessResponse{data=dto.TokenResponse}
// @Failure      400 {object} response.ErrorResponse{error=apperror.AppError} "Validation error"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Wrong email or password"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /login [post]
func (h *UserHandlerImpl) Login(ctx *gin.Context) {
	var request dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	if err := h.validator.Struct(&request); err != nil {
		response.HandleError(ctx, apperror.ExtractValidationErrors(err))
		return
	}

	token, err := h.userUsecase.Login(ctx.Request.Context(), &request)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, token)
}

// GetProfile godoc
// @Summary      Get the authenticated user's profile
// @Description  Returns the profile of the user identified by the JWT in the Authorization header. The password hash is never included in the response.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      404 {object} response.ErrorResponse{error=apperror.AppError} "User not found"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /profile [get]
func (h *UserHandlerImpl) GetProfile(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.HandleError(ctx, apperror.ErrUnauthorized)
		return
	}

	user, err := h.userUsecase.GetProfile(ctx.Request.Context(), auth.ID)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.ResponseSuccess(ctx, http.StatusOK, user)
}

// Logout godoc
// @Summary      Logout a user
// @Description  Logs out a user.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse "User logged out successfully"
// @Failure      401 {object} response.ErrorResponse{error=apperror.AppError} "Unauthorized"
// @Failure      500 {object} response.ErrorResponse{error=apperror.AppError} "Internal server error"
// @Router       /logout [post]
func (h *UserHandlerImpl) Logout(ctx *gin.Context) {
	token := ctx.GetString("token")
	if token == "" {
		response.HandleError(ctx, apperror.ErrBadRequest)
		return
	}

	err := h.userUsecase.Logout(ctx.Request.Context(), token)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	response.ResponseSuccess(ctx, http.StatusOK, nil)
}
