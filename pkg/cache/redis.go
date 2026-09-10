package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func NewRedisCli(cfg *config.Config, log *logger.Logger) (*redis.Client, func(), error) {
	var opts *redis.Options
	var err error
	if cfg.RedisURL != "" {
		opts, err = redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse redis url: %w", err)
		}
	} else {
		addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
		opts = &redis.Options{
			Addr:     addr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}

		if cfg.RedisTLSEnabled {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
	}

	client := redis.NewClient(opts)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info("Connected to Redis successfully")

	cleanup := func() {
		if err := client.Close(); err != nil {
			log.Errorf("failed to close redis client: %v", err)
		}
		log.Info("Redis connection closed")
	}

	return client, cleanup, nil
}
