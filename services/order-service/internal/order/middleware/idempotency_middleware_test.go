package middleware_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/response"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/dto"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/middleware"
	"github.com/Mpayy/e-commerce/services/order-service/internal/order/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func setupIdempotencyTest(t *testing.T) (*gin.Engine, *mocks.MockIdempotencyRepository) {
	gin.SetMode(gin.TestMode)

	idemRepo := mocks.NewMockIdempotencyRepository(t)
	log := newTestLogger()

	t.Cleanup(func() {
		idemRepo.AssertExpectations(t)
	})

	// Variable counter untuk menandai eksekusi handler asli
	handlerCallCount := 0

	router := gin.New()
	router.POST("/orders", middleware.IdempotencyMiddleware(idemRepo, log), func(c *gin.Context) {
		handlerCallCount++
		c.JSON(http.StatusCreated, response.SuccessResponse{
			Success: true,
			Data: dto.OrderResponse{
				OrderID:       uint(handlerCallCount),
				InvoiceNumber: fmt.Sprintf("INV-HANDLER-%d", handlerCallCount),
				TotalAmount:   40000,
				Status:        "PAID",
			},
		})
	})

	return router, idemRepo
}

func TestIdempotencyMiddleware(t *testing.T) {
	t.Run("success_first_request", func(t *testing.T) {
		router, idemRepo := setupIdempotencyTest(t)
		idemKey := "test-uuid-1"

		idemRepo.EXPECT().Lock(mock.Anything, idemKey, 30*time.Second).Return(true, nil)
		idemRepo.EXPECT().SaveResponse(mock.Anything, idemKey, http.StatusCreated, mock.Anything, 24*time.Hour).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", idemKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "INV-HANDLER-1")
	})

	t.Run("success_retry_cached_response", func(t *testing.T) {
		router, idemRepo := setupIdempotencyTest(t)
		idemKey := "test-uuid-1"

		// Response dari cache sengaja menggunakan ID/Invoice unik "FROM-CACHE"
		cachedResponse := &dto.IdempotencyResponse{
			StatusCode: http.StatusCreated,
			Body: &dto.OrderResponse{
				OrderID:       999,
				InvoiceNumber: "INV-FROM-CACHE-999",
				TotalAmount:   40000,
				Status:        "PAID",
			},
		}

		idemRepo.EXPECT().Lock(mock.Anything, idemKey, 30*time.Second).Return(false, nil)
		idemRepo.EXPECT().GetResponse(mock.Anything, idemKey).Return(cachedResponse, nil)

		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", idemKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		// Memastikan response berasal dari CACHE, bukan dari handler ("INV-HANDLER-1")
		assert.Contains(t, w.Body.String(), "INV-FROM-CACHE-999")
	})

	t.Run("error_concurrent_request_locked", func(t *testing.T) {
		router, idemRepo := setupIdempotencyTest(t)
		idemKey := "test-uuid-concurrent"

		idemRepo.EXPECT().Lock(mock.Anything, idemKey, 30*time.Second).Return(false, nil)
		idemRepo.EXPECT().GetResponse(mock.Anything, idemKey).Return(nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", idemKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("success_key_expired_processed_as_new", func(t *testing.T) {
		router, idemRepo := setupIdempotencyTest(t)
		idemKey := "test-uuid-expired"

		idemRepo.EXPECT().Lock(mock.Anything, idemKey, 30*time.Second).Return(true, nil)
		idemRepo.EXPECT().SaveResponse(mock.Anything, idemKey, http.StatusCreated, mock.Anything, 24*time.Hour).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", idemKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "INV-HANDLER-1")
	})
}
