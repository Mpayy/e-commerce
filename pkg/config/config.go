package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// App Configuration
	AppHost  string `mapstructure:"APP_HOST"`
	AppPort  string `mapstructure:"APP_PORT"`
	AppEnv   string `mapstructure:"APP_ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// Relational Database Configuration (Postgres / MySQL)
	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     string `mapstructure:"DATABASE_PORT"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabaseUsername string `mapstructure:"DATABASE_USERNAME"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseSSLMode  string `mapstructure:"DATABASE_SSL_MODE"`

	// Redis Configuration
	RedisHost       string `mapstructure:"REDIS_HOST"`
	RedisPort       string `mapstructure:"REDIS_PORT"`
	RedisPassword   string `mapstructure:"REDIS_PASSWORD"`
	RedisDB         int    `mapstructure:"REDIS_DB"`
	RedisTLSEnabled bool   `mapstructure:"REDIS_TLS_ENABLED"`

	// MongoDB Configuration
	MongoURI string `mapstructure:"MONGODB_URI"`
	MongoDB  string `mapstructure:"MONGODB_DATABASE"`

	// Auth Configuration
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`

	// GRPC
	ProductServiceAddr string `mapstructure:"PRODUCT_SERVICE_ADDR"`
}

func Load() *Config {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	// Set Default Values (fallback jika di .env tidak diisi)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_HOST", "localhost")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")

	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_PORT", "5432")
	v.SetDefault("DATABASE_NAME", "ecommerce")
	v.SetDefault("DATABASE_USERNAME", "postgres")
	v.SetDefault("DATABASE_PASSWORD", "postgres")
	v.SetDefault("DATABASE_SSL_MODE", "disable")

	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_TLS_ENABLED", false)

	v.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	v.SetDefault("MONGODB_DATABASE", "ecommerce")

	v.SetDefault("PRODUCT_SERVICE_ADDR", "localhost:50051")

	// Otomatis override jika ada Environment Variable di OS / Docker
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: Config file error: %v", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &cfg
}
