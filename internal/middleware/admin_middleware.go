package middleware

import (
	"net/http"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminMiddleware struct{}

func NewAdminMiddleware() *AdminMiddleware {
	return &AdminMiddleware{}
}

func (m *AdminMiddleware) AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authValue, exists := ctx.Get("auth")
		if !exists {
			response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error(), nil)
			return
		}

		auth, ok := authValue.(*jwt.Auth)
		if !ok {
			response.ResponseError(ctx, http.StatusInternalServerError, apperror.ErrInternalServer.Error(), nil)
			return
		}

		if auth.Role != "admin" {
			response.ResponseError(ctx, http.StatusForbidden, apperror.ErrForbidden.Error(), nil)
			return
		}

		ctx.Next()
	}
}