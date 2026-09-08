package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/middleware"
	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/mocks"
)

func setupRateLimiterTest(t *testing.T) (*middleware.RateLimiter, *mocks.MockHandler) {
	mockNext := mocks.NewMockHandler(t)
	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(func() {
		cancel()
		mockNext.AssertExpectations(t)
	})

	rateLimiter := middleware.NewRateLimiter(ctx)
	return rateLimiter, mockNext
}

func TestRateLimiter_Middleware(t *testing.T) {
	t.Run("success_strict_limit_login_route", func(t *testing.T) {
		rl, mockNext := setupRateLimiterTest(t)

		mockNext.EXPECT().
			ServeHTTP(mock.Anything, mock.Anything).
			Run(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}).
			Once()

		handlerToTest := rl.Middleware(mockNext)

		// Menggunakan route /api/v1/login sesuai logika middleware
		req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
		req.RemoteAddr = "203.0.113.195:12345"
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "5", rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("success_strict_limit_register_route", func(t *testing.T) {
		rl, mockNext := setupRateLimiterTest(t)

		mockNext.EXPECT().
			ServeHTTP(mock.Anything, mock.Anything).
			Run(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}).
			Once()

		handlerToTest := rl.Middleware(mockNext)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/register", nil)
		req.RemoteAddr = "192.168.1.50:54321"
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "5", rec.Header().Get("X-RateLimit-Limit"))
	})

	t.Run("success_fallback_ip_when_remote_addr_has_no_port", func(t *testing.T) {
		rl, mockNext := setupRateLimiterTest(t)

		mockNext.EXPECT().
			ServeHTTP(mock.Anything, mock.Anything).
			Run(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}).
			Once()

		handlerToTest := rl.Middleware(mockNext)

		// RemoteAddr tanpa port untuk menguji err != nil pada net.SplitHostPort
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		req.RemoteAddr = "192.168.1.100"
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "60", rec.Header().Get("X-RateLimit-Limit"))
	})

	t.Run("success_general_route_limit_and_cache_hit_get_limiter", func(t *testing.T) {
		rl, mockNext := setupRateLimiterTest(t)

		mockNext.EXPECT().
			ServeHTTP(mock.Anything, mock.Anything).
			Run(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}).
			Times(2)

		handlerToTest := rl.Middleware(mockNext)
		clientIP := "10.0.0.1:12345"

		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		req1.RemoteAddr = clientIP
		rec1 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec1, req1)

		assert.Equal(t, http.StatusOK, rec1.Code)
		assert.Equal(t, "60", rec1.Header().Get("X-RateLimit-Limit"))

		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		req2.RemoteAddr = clientIP
		rec2 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusOK, rec2.Code)
		assert.Equal(t, "60", rec2.Header().Get("X-RateLimit-Limit"))
	})

	t.Run("error_rate_limit_exceeded_status_429", func(t *testing.T) {
		rl, mockNext := setupRateLimiterTest(t)

		mockNext.EXPECT().
			ServeHTTP(mock.Anything, mock.Anything).
			Run(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}).
			Times(5)

		handlerToTest := rl.Middleware(mockNext)
		clientIP := "172.16.0.1:33333"

		// Kirim 5 request pertama (batas burst untuk /api/v1/login)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
			req.RemoteAddr = clientIP
			rec := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Request ke-6 harus gagal HTTP 429
		reqSpam := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
		reqSpam.RemoteAddr = clientIP
		recSpam := httptest.NewRecorder()

		handlerToTest.ServeHTTP(recSpam, reqSpam)

		assert.Equal(t, http.StatusTooManyRequests, recSpam.Code)
		assert.Equal(t, "application/json", recSpam.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"error": "Too Many Requests"}`, recSpam.Body.String())
	})
}
