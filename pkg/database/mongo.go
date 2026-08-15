package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoDB(cfg *config.Config, log *logger.Logger) (*mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := cfg.MongoURI
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	log.Info("Connected to MongoDB successfully")

	cleanup := func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()

		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Errorf("failed to disconnect mongodb: %v", err)
		} else {
			log.Info("MongoDB connection closed")
		}
	}

	dbName := cfg.MongoDB
	if dbName == "" {
		dbName = "ecommerce"
	}

	db := client.Database(dbName)

	return db, cleanup, nil
}