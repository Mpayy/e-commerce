package middleware

import (
	"net/http"
	"strings"

	"github.com/Mpayy/e-commerce/dependency"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	TokenUtil jwt.JwtToken
	Redis     dependency.Redis
}

func NewAuthMiddleware(tokenUtil jwt.JwtToken, redisClient dependency.Redis) *AuthMiddleware {
	return &AuthMiddleware{TokenUtil: tokenUtil, Redis: redisClient}
}

func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" || tokenString == "Bearer" {
			response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error() , nil)
			return
		}

		auth, err := m.TokenUtil.ParseToken(tokenString)
		if err != nil {
			response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error() , nil)
			return
		}

		exists, err := m.Redis.CheckToRedis(ctx.Request.Context(), dependency.AuthPrefix+tokenString)
		if err != nil || !exists {
			response.ResponseError(ctx, http.StatusUnauthorized, apperror.ErrUnauthorized.Error() , nil)
			return
		}

		ctx.Set("auth", auth)
		ctx.Set("token", tokenString)

		ctx.Next()
	}
}
	
func GetAuthUser(ctx *gin.Context) *jwt.Auth {
	authValue, exists := ctx.Get("auth")
	if !exists {
		return nil
	}
	
	auth, ok := authValue.(*jwt.Auth)
	if !ok {
		return nil
	}
	
	return auth
}