package dependency

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoDatabase(client *mongo.Client, config *viper.Viper) *mongo.Database {
	dbName := config.GetString("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "ecommerce"
	}
	db := client.Database(dbName)
	return db
}

func NewMongoClient(config *viper.Viper, log *logrus.Logger) (*mongo.Client, func(), error) {
	uri := config.GetString("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to ping mongodb at %s: %w", uri, err)
	}

	log.Info("mongodb connected successfully")

	cleanup := func() {
		ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()

		if err := client.Disconnect(ctx); err != nil {
			log.WithError(err).Error("error closing mongodb connection")
		}
		log.Info("mongodb connection closed")
	}

	return client, cleanup, nil
}
