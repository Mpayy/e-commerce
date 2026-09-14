package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/repository"
	"github.com/gin-gonic/gin"
)

const (
	lockTTL   = 30 * time.Second
	resultTTL = 24 * time.Hour
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func IdempotencyMiddleware(repo repository.IdempotencyRepository, log *logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idemKey := ctx.GetHeader("Idempotency-Key")
		if idemKey == "" {
			response.HandleError(ctx, apperror.ErrBadRequest)
			return
		}

		locked, err := repo.Lock(ctx.Request.Context(), idemKey, lockTTL)
		if err != nil {
			response.HandleError(ctx, fmt.Errorf("failed to lock idempotency: %w", err))
			return
		}

		if !locked {
			cache, err := repo.GetResponse(ctx.Request.Context(), idemKey)
			if err != nil {
				response.HandleError(ctx, fmt.Errorf("get idempotency response: %w", err))
				return
			}

			if cache != nil {
				response.ResponseSuccess(ctx, cache.StatusCode, cache.Body)
				ctx.Abort()
				return
			}

			response.HandleError(ctx, apperror.ErrIdempotencyKeyLocked)
			return
		}

		w := &responseBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBufferString(""),
		}
		ctx.Writer = w

		ctx.Next()

		statusCode := ctx.Writer.Status()
		capturedBody := w.body.Bytes()

		if statusCode >= 200 && statusCode < 300 {
			var wrappedResp struct {
				Success bool              `json:"success"`
				Data    dto.OrderResponse `json:"data"`
			}

			if err := json.Unmarshal(capturedBody, &wrappedResp); err != nil {
				log.WithError(err).Error("failed to unmarshal response for idempotency caching")
				return
			}

			if err := repo.SaveResponse(ctx.Request.Context(), idemKey, statusCode, &wrappedResp.Data, resultTTL); err != nil {
				log.WithError(err).Error("failed to save idempotency response")
				return
			}
		} else {
			if err := repo.Unlock(ctx.Request.Context(), idemKey); err != nil {
				log.WithError(err).Error("failed to unlock idempotency")
				return
			}
		}
	}
}
