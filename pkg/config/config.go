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
	AppUrl   string `mapstructure:"APP_URL"`
	AppEnv   string `mapstructure:"APP_ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// PostgreSQL Configuration
	DatabaseURL      string `mapstructure:"DATABASE_URL"`
	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     string `mapstructure:"DATABASE_PORT"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabaseUsername string `mapstructure:"DATABASE_USERNAME"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseSSLMode  string `mapstructure:"DATABASE_SSLMODE"`

	// Redis Configuration
	RedisURL        string `mapstructure:"REDIS_URL"`
	RedisHost       string `mapstructure:"REDIS_HOST"`
	RedisPort       string `mapstructure:"REDIS_PORT"`
	RedisPassword   string `mapstructure:"REDIS_PASSWORD"`
	RedisDB         int    `mapstructure:"REDIS_DB"`
	RedisTLSEnabled bool   `mapstructure:"REDIS_TLS_ENABLED"`

	// MongoDB Configuration
	MongoURL          string `mapstructure:"MONGODB_URL"`
	MongoHost         string `mapstructure:"MONGODB_HOST"`
	MongoPort         string `mapstructure:"MONGODB_PORT"`
	MongoDB           string `mapstructure:"MONGODB_DATABASE"`
	MongoDBReplicaSet string `mapstructure:"MONGODB_REPLICASET"`

	// RabbitMQ Configuration
	RabbitMQURL      string `mapstructure:"RABBITMQ_URL"`
	RabbitMQUser     string `mapstructure:"RABBITMQ_USER"`
	RabbitMQPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitMQHost     string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQPort     string `mapstructure:"RABBITMQ_PORT"`

	// Auth Configuration
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`

	// GRPC
	ProductServiceAddr string `mapstructure:"PRODUCT_SERVICE_ADDR"`

	// Gateway
	UserServiceAddr        string `mapstructure:"USER_SERVICE_ADDR"`
	ProductServiceHTTPAddr string `mapstructure:"PRODUCT_SERVICE_HTTP_ADDR"`
	OrderServiceAddr       string `mapstructure:"ORDER_SERVICE_ADDR"`

	// Image Upload
	BasePath             string `mapstructure:"BASE_PATH"`
	PublicImageURLPrefix string `mapstructure:"PUBLIC_IMAGE_URL_PREFIX"`
}

func Load() *Config {
	v := viper.New()

	v.SetConfigFile(".env")

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_HOST", "localhost")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_URL", "http://localhost:8080")
	v.SetDefault("LOG_LEVEL", "info")

	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_PORT", "5432")
	v.SetDefault("DATABASE_NAME", "ecommerce")
	v.SetDefault("DATABASE_USERNAME", "postgres")
	v.SetDefault("DATABASE_PASSWORD", "postgres")
	v.SetDefault("DATABASE_SSLMODE", "disable")

	v.SetDefault("REDIS_URL", "")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_TLS_ENABLED", false)

	v.SetDefault("MONGODB_URL", "")
	v.SetDefault("MONGODB_HOST", "localhost")
	v.SetDefault("MONGODB_PORT", "27017")
	v.SetDefault("MONGODB_DATABASE", "ecommerce")
	v.SetDefault("MONGODB_REPLICASET", "rs0")

	v.SetDefault("PRODUCT_SERVICE_ADDR", "localhost:50051")

	v.SetDefault("RABBITMQ_URL", "")
	v.SetDefault("RABBITMQ_USER", "guest")
	v.SetDefault("RABBITMQ_PASSWORD", "guest")
	v.SetDefault("RABBITMQ_HOST", "localhost")
	v.SetDefault("RABBITMQ_PORT", "5672")

	v.SetDefault("USER_SERVICE_ADDR", "http://user-service:8082")
	v.SetDefault("PRODUCT_SERVICE_HTTP_ADDR", "http://product-service:8081")
	v.SetDefault("ORDER_SERVICE_ADDR", "http://order-service:8083")

	v.SetDefault("BASE_PATH", "./uploads/products")
	v.SetDefault("PUBLIC_IMAGE_URL_PREFIX", "/uploads/products")

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
