package middleware

import (
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authValue, exists := ctx.Get("auth")
		if !exists {
			response.HandleError(ctx, apperror.ErrUnauthorized)
			return
		}

		auth, ok := authValue.(*jwt.Auth)
		if !ok {
			response.HandleError(ctx, apperror.ErrInternalServer)
			return
		}

		if auth.Role != "admin" {
			response.HandleError(ctx, apperror.ErrForbidden)
			return
		}

		ctx.Next()
	}
}
