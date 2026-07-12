package userhttp

import (
	"errors"
	"net/http"

	"github.com/Mpayy/e-commerce/internal/middleware"
	"github.com/Mpayy/e-commerce/internal/user/dto"
	userusecase "github.com/Mpayy/e-commerce/internal/user/usecase"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type UserHandlerImpl struct {
	UserUsecase userusecase.UserUsecase
	Validator   *validator.Validate
	Log         *logrus.Logger
}

func NewUserHandler(userUsecase userusecase.UserUsecase, validator *validator.Validate, log *logrus.Logger) UserHandler {
	return &UserHandlerImpl{
		UserUsecase: userUsecase,
		Validator:   validator,
		Log:         log,
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
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      409 {object} response.ErrorResponse "Email already exists"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /register [post]
func (h *UserHandlerImpl) Register(ctx *gin.Context) {
	var request dto.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		h.Log.WithError(err).Error("Validation error during registration")
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	user, err := h.UserUsecase.Register(ctx.Request.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrDuplicatedEmail):
			response.ResponseError(ctx, http.StatusConflict, err.Error(), nil)
			return
		default:
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}
	}

	response.ResponseSuccess(ctx, http.StatusCreated, "user registered successfully", user)
}

// Login godoc
// @Summary      Login and obtain a JWT token
// @Description  Authenticates a user by email and password, then issues a JWT stored in Redis for session tracking. Returns a generic error for both wrong password and unknown email to avoid leaking which emails are registered.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.UserLoginRequest true "User payload"
// @Success      200 {object} response.SuccessResponse{data=dto.TokenResponse}
// @Failure      400 {object} response.ValidationErrorResponse "Validation error"
// @Failure      401 {object} response.ErrorResponse "Wrong email or password"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /login [post]
func (h *UserHandlerImpl) Login(ctx *gin.Context) {
	var request dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	if err := h.Validator.Struct(&request); err != nil {
		errorReport := apperror.ExtractValidationErrors(err)
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrValidationFailed.Error(), errorReport)
		return
	}

	token, err := h.UserUsecase.Login(ctx.Request.Context(), &request)
	if err != nil {
		if errors.Is(err, apperror.ErrWrongEmailOrPassword) {
			response.ResponseError(ctx, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		h.Log.WithError(err).Error("Unexpected error during login")
		response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.ResponseSuccess(ctx, http.StatusOK, "user logged in successfully", token)
}

// GetProfile godoc
// @Summary      Get the authenticated user's profile
// @Description  Returns the profile of the user identified by the JWT in the Authorization header. The password hash is never included in the response.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      404 {object} response.ErrorResponse "User not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /profile [get]
func (h *UserHandlerImpl) GetProfile(ctx *gin.Context) {
	auth := middleware.GetAuthUser(ctx)
	if auth == nil {
		response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
		return
	}

	user, err := h.UserUsecase.GetProfile(ctx.Request.Context(), auth.UserID)
	if err != nil {
		if errors.Is(err, apperror.ErrUserNotFound) {
			response.ResponseError(ctx, http.StatusNotFound, err.Error(), nil)
			return
		}
		h.Log.WithError(err).Error("Unexpected error during get profile")
		response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.ResponseSuccess(ctx, http.StatusOK, "user profile retrieved successfully", user)
}

// Logout godoc
// @Summary      Logout a user
// @Description  Logs out a user.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.SuccessResponse "User logged out successfully"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /logout [post]
func (h *UserHandlerImpl) Logout(ctx *gin.Context) {
	token := ctx.GetString("token")
	if token == "" {
		response.ResponseError(ctx, http.StatusBadRequest, apperror.ErrBadRequest.Error(), nil)
		return
	}

	err := h.UserUsecase.Logout(ctx.Request.Context(), token)
	if err != nil {
		h.Log.WithError(err).Error("Unexpected error during logout")
		response.ResponseError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.ResponseSuccess(ctx, http.StatusOK, "user logged out successfully", nil)
}
