package response

import "github.com/gin-gonic/gin"

type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

type ValidationErrorResponse struct {
	Success bool              `json:"success" example:"false"`
	Message string            `json:"message" example:"Validation failed"`
	Error   map[string]string `json:"error" example:"field_name:error message,quantity:must be at least 1"`
}

func ResponseSuccess(ctx *gin.Context, code int, message string, data any) {
	ctx.JSON(code, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ResponseError(ctx *gin.Context, code int, message string, err any) {
	ctx.AbortWithStatusJSON(code, ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}
