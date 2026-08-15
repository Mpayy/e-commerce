package cache

import (
	"crypto/tls"
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func NewRedisCli(cfg *config.Config, log *logger.Logger) (*redis.Client, func()) {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	password := cfg.RedisPassword
	db := cfg.RedisDB
	enableTLS := cfg.RedisTLSEnabled

	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	if enableTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)

	log.Info("Connected to Redis successfully")

	cleanup := func() {
		if err := client.Close(); err != nil {
			log.Errorf("failed to close redis client: %v", err)
		}
		log.Info("Redis connection closed")
	}

	return client, cleanup
}