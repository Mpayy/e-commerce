package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionChecker interface {
	SessionExists(ctx context.Context, token string) (bool, error)
}

type AuthMiddleware struct {
	jwtToken    jwt.JwtToken
	sessionRepo SessionChecker
	log         *logger.Logger
}

func NewAuthMiddleware(jwtToken jwt.JwtToken, sessionRepo SessionChecker, log *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwtToken: jwtToken, sessionRepo: sessionRepo, log: log}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.HandleError(ctx, apperror.ErrUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" || token == "Bearer" {
			response.HandleError(ctx, apperror.ErrUnauthorized)
			return
		}

		auth, err := m.jwtToken.Validate(token)
		if err != nil {
			m.log.WithError(err).Debug("jwt validation failed")
			response.HandleError(ctx, err)
			return
		}

		exists, err := m.sessionRepo.SessionExists(ctx, token)
		if err != nil {
			response.HandleError(ctx, fmt.Errorf("check session: %w", err))
			return
		}
		if !exists {
			response.HandleError(ctx, apperror.ErrUnauthorized)
			return
		}

		ctx.Set("auth", auth)
		ctx.Set("token", token)

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
