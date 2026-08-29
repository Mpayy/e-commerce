package apperror

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AppError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	Status  int               `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

var (
	// Common / General Errors
	ErrInternalServer      = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "something went wrong", Status: http.StatusInternalServerError}
	ErrDuplicatedKey       = &AppError{Code: "DUPLICATED_KEY", Message: "duplicated key", Status: http.StatusConflict}
	ErrRecordNotFound      = &AppError{Code: "RECORD_NOT_FOUND", Message: "record not found", Status: http.StatusNotFound}
	ErrInvalidID           = &AppError{Code: "INVALID_ID", Message: "invalid id", Status: http.StatusBadRequest}
	ErrBadRequest          = &AppError{Code: "BAD_REQUEST", Message: "bad request", Status: http.StatusBadRequest}
	ErrNoActiveTransaction = &AppError{Code: "NO_ACTIVE_TRANSACTION", Message: "no active transaction", Status: http.StatusInternalServerError}
	ErrValidationFailed    = &AppError{Code: "VALIDATION_FAILED", Message: "validation failed", Status: http.StatusUnprocessableEntity}

	// Auth & User Errors
	ErrUnauthorized         = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized access", Status: http.StatusUnauthorized}
	ErrForbidden            = &AppError{Code: "FORBIDDEN", Message: "access forbidden", Status: http.StatusForbidden}
	ErrWrongEmailOrPassword = &AppError{Code: "WRONG_EMAIL_OR_PASSWORD", Message: "wrong email or password", Status: http.StatusUnauthorized}
	ErrDuplicatedEmail      = &AppError{Code: "DUPLICATED_EMAIL", Message: "email already exists", Status: http.StatusConflict}
	ErrUserNotFound         = &AppError{Code: "USER_NOT_FOUND", Message: "user not found", Status: http.StatusNotFound}
	ErrExpiredToken         = &AppError{Code: "EXPIRED_TOKEN", Message: "token expired", Status: http.StatusUnauthorized}
	ErrInvalidToken         = &AppError{Code: "INVALID_TOKEN", Message: "invalid token", Status: http.StatusUnauthorized}

	// Category Errors
	ErrCategoryNotFound   = &AppError{Code: "CATEGORY_NOT_FOUND", Message: "category not found", Status: http.StatusNotFound}
	ErrDuplicatedCategory = &AppError{Code: "DUPLICATED_CATEGORY", Message: "duplicated category", Status: http.StatusConflict}

	// Product Errors
	ErrProductNotFound      = &AppError{Code: "PRODUCT_NOT_FOUND", Message: "product not found", Status: http.StatusNotFound}
	ErrDuplicatedProduct    = &AppError{Code: "DUPLICATED_PRODUCT", Message: "duplicated product", Status: http.StatusConflict}
	ErrDuplicatedProductSku = &AppError{Code: "DUPLICATED_PRODUCT_SKU", Message: "duplicated product SKU", Status: http.StatusConflict}
	ErrInsufficientStock    = &AppError{Code: "INSUFFICIENT_STOCK", Message: "insufficient stock", Status: http.StatusConflict}

	// Cart Errors
	ErrCartNotFound    = &AppError{Code: "CART_NOT_FOUND", Message: "cart not found", Status: http.StatusNotFound}
	ErrCartEmpty       = &AppError{Code: "CART_EMPTY", Message: "cart is empty", Status: http.StatusBadRequest}
	ErrInvalidQuantity = &AppError{Code: "INVALID_QUANTITY", Message: "invalid quantity", Status: http.StatusBadRequest}

	// Order Errors
	ErrOrderNotFound    = &AppError{Code: "ORDER_NOT_FOUND", Message: "order not found", Status: http.StatusNotFound}
	ErrInvalidDateRange = &AppError{Code: "INVALID_DATE_RANGE", Message: "invalid date range", Status: http.StatusBadRequest}
)

func ExtractValidationErrors(err error) *AppError {
	errorReport := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorReport = TranslateValidationError(validationErrors)
	}

	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: "one or more fields are invalid",
		Fields:  errorReport,
		Status:  http.StatusBadRequest,
	}
}

func TranslateValidationError(valErr validator.ValidationErrors) map[string]string {
	fieldError := make(map[string]string)

	for _, e := range valErr {
		fieldError[strings.ToLower(e.Field())] = translateTag(e)
	}

	return fieldError
}

func translateTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "must be filled"
	case "min":
		return "must be at least " + e.Param()
	case "max":
		return "must be at most " + e.Param()
	case "gt":
		return "must be greater than " + e.Param()
	case "gte":
		return "must be greater than or equal to " + e.Param()
	default:
		return "invalid input value"
	}
}
