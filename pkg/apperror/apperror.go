package apperror

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrDuplicatedKey        = errors.New("Email Already Exists")
	ErrNotFound             = errors.New("Data Not Found")
	ErrUnauthorized         = errors.New("Unauthorized")
	ErrForbidden            = errors.New("Forbidden")
	ErrInternalServer       = errors.New("Internal Server Error")
	ErrBadRequest           = errors.New("Bad Request")
	ErrWrongEmailOrPassword = errors.New("Wrong Email or Password")
	ErrValidationFailed     = errors.New("Validation Failed")
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
