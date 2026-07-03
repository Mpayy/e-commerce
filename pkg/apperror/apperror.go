package apperror

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// Error Duplicated
	ErrDuplicatedKey        = errors.New("Data Already Exists")
	ErrDuplicatedEmail      = errors.New("Email Already Exists")
	ErrDuplicatedCategory   = errors.New("Category Already Exists")
	ErrDuplicatedProduct    = errors.New("Product Already Exists")
	ErrDuplicatedProductSku = errors.New("Product SKU Already Exists")

	// Error Not Found
	ErrNotFound         = errors.New("Data Not Found")
	ErrCategoryNotFound = errors.New("Category Not Found")
	ErrProductNotFound  = errors.New("Product Not Found")

	// Error
	ErrUnauthorized         = errors.New("Unauthorized")
	ErrForbidden            = errors.New("Forbidden")
	ErrInternalServer       = errors.New("Internal Server Error")
	ErrBadRequest           = errors.New("Bad Request")
	ErrWrongEmailOrPassword = errors.New("Wrong Email or Password")
	ErrValidationFailed     = errors.New("Validation Failed")
	ErrInsufficientStock    = errors.New("Insufficient Stock")
)

func ExtractValidationErrors(err error) map[string]string {
	errorReport := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorReport = TranslateValidationError(validationErrors)
	}

	return errorReport
}

func TranslateValidationError(valErr validator.ValidationErrors) map[string]string {
	fieldError := make(map[string]string)

	for _, e := range valErr {
		var message string
		switch e.Tag() {
		case "required":
			message = "must be filled"
		case "email":
			message = "must be a valid email"
		case "min":
			message = "must be at least " + e.Param() + " characters long"
		case "max":
			message = "must be at most " + e.Param() + " characters long"
		default:
			message = "invalid input value"
		}
		fieldError[strings.ToLower(e.Field())] = message
	}

	return fieldError
}
