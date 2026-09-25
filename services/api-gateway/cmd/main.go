package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	gwConfig "github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/config"
	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/middleware"
	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/proxy"
	"github.com/rs/cors"
)

func getAllowedOrigins(cfg *config.Config) []string {
	rawOrigin := cfg.CorsAllowedOrigins

	if rawOrigin == "" {
		return []string{"http://localhost:3000"}
	}

	origins := strings.Split(rawOrigin, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.NewLogger(cfg)
	rateLimiter := middleware.NewRateLimiter(ctx)

	targets := gwConfig.ServiceTargets{
		UserServiceAddr:    cfg.UserServiceAddr,
		ProductServiceAddr: cfg.ProductServiceHTTPAddr,
		OrderServiceAddr:   cfg.OrderServiceAddr,
	}

	gateway, err := proxy.NewGateway(gwConfig.BuildRoutes(targets))
	if err != nil {
		log.Fatalf("failed to initialize gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})
	mux.Handle("/", rateLimiter.Middleware(gateway.Handler()))

	c := cors.New(cors.Options{
		AllowedOrigins:   getAllowedOrigins(cfg),
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"authorization", "content-type", "idempotency-key"},
		AllowCredentials: false,
		MaxAge:           int((12 * time.Hour).Seconds()),
	})

	handlerWithCORS := c.Handler(mux)

	srv := &http.Server{Addr: ":8080", Handler: handlerWithCORS}

	go func() {
		log.Infof("API Gateway running on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gateway server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("API Gateway shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")
}
