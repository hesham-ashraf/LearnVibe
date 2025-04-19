package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Server settings
	Port string

	// JWT settings
	JWTSecret string

	// Database settings
	DatabaseURL string

	// Redis (caching) settings
	RedisURL  string
	RedisPass string

	// MinIO (object storage) settings
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	// CMS Service URL for inter-service communication
	CMSServiceURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load()

	// Load configuration
	config := &Config{
		Port:           getEnv("PORT", "8082"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:vampire8122003@localhost:5432/learnvibe_content"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		RedisPass:      getEnv("REDIS_PASSWORD", ""),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "learnvibe-content"),
		MinioUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		CMSServiceURL:  getEnv("CMS_SERVICE_URL", "http://localhost:8080"),
	}

	// Validate required settings
	if config.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET is not set! Using default value for development only!")
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool gets an environment variable as boolean or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}