package dependency

import (
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewProductServiceConn(cfg *config.Config, log *logger.Logger) (*grpc.ClientConn, func(), error) {
	addr := cfg.ProductServiceAddr
	if addr == "" {
		addr = "product-service:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	log.Info("Connected to product service successfully")

	cleanup := func() {
		err := conn.Close()
		if err != nil {
			log.Errorf("failed to close product service connection: %v", err)
		}
		log.Info("Product service connection closed")
	}

	return conn, cleanup, nil
}
