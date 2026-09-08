package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
}

func NewRateLimiter(ctx context.Context) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
	}
	go rl.cleanupStaleClients(ctx)
	return rl
}

func (rl *RateLimiter) cleanupStaleClients(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			for key, c := range rl.clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(rl.clients, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) GetLimiter(ip string, limit rate.Limit, burst int) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(limit, burst)
		rl.clients[ip] = &client{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		var limit rate.Limit
		var burst int

		category := "general"
		limit = rate.Every(time.Second)
		burst = 60

		if r.URL.Path == "/api/v1/login" || r.URL.Path == "/api/v1/register" {
			category = "strict"
			limit = rate.Every(12 * time.Second)
			burst = 5
		}

		limiter := rl.GetLimiter(ip+"-"+category, limit, burst)

		allowed := limiter.Allow()
		tokens := int(limiter.Tokens())
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(burst))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(tokens))

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "Too Many Requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
