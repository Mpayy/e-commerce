package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	gwConfig "github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/config"
	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/proxy"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.NewLogger(cfg)

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
	mux.Handle("/", gateway.Handler())

	srv := &http.Server{Addr: ":8080", Handler: mux}

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
