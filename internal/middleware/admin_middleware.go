package middleware

import (
	"net/http"

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
			response.ResponseError(ctx, http.StatusUnauthorized, "unauthorized", nil)
			return
		}

		auth, ok := authValue.(*jwt.Auth) 
		if !ok {
			response.ResponseError(ctx, http.StatusInternalServerError, "internal server error", nil)
			return
		}

		if auth.Role != "admin" {
			response.ResponseError(ctx, http.StatusForbidden, "admin access required", nil)
			return
		}

		ctx.Next()
	}
}