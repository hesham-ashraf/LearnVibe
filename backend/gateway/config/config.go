package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Server settings
	Port        string
	EnableHTTPS bool
	HTTPSPort   string

	// JWT settings for validation
	JWTSecret string

	// Rate limiting
	RateLimitRequests int
	RateLimitDuration int // in seconds

	// Circuit breaker settings
	CircuitBreakerMaxRequests uint32
	CircuitBreakerInterval    int // in seconds
	CircuitBreakerTimeout     int // in seconds

	// Service URLs
	CMSServiceURL         string
	ContentServiceURL     string
	CMSHealthEndpoint     string
	ContentHealthEndpoint string

	// Request timeouts
	RequestTimeout int // in seconds
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load()

	// Load configuration
	config := &Config{
		Port:                      getEnv("PORT", "8000"),
		EnableHTTPS:               getEnvAsBool("ENABLE_HTTPS", false),
		HTTPSPort:                 getEnv("HTTPS_PORT", "8443"),
		JWTSecret:                 getEnv("JWT_SECRET", "your-secret-key"),
		RateLimitRequests:         getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration:         getEnvAsInt("RATE_LIMIT_DURATION", 60),
		CircuitBreakerMaxRequests: uint32(getEnvAsInt("CIRCUIT_BREAKER_MAX_REQUESTS", 5)),
		CircuitBreakerInterval:    getEnvAsInt("CIRCUIT_BREAKER_INTERVAL", 30),
		CircuitBreakerTimeout:     getEnvAsInt("CIRCUIT_BREAKER_TIMEOUT", 10),
		CMSServiceURL:             getEnv("CMS_SERVICE_URL", "http://localhost:8080"),
		ContentServiceURL:         getEnv("CONTENT_SERVICE_URL", "http://localhost:8082"),
		CMSHealthEndpoint:         getEnv("CMS_HEALTH_ENDPOINT", "/health"),
		ContentHealthEndpoint:     getEnv("CONTENT_HEALTH_ENDPOINT", "/health"),
		RequestTimeout:            getEnvAsInt("REQUEST_TIMEOUT", 30),
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

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("WARNING: Invalid integer format for %s, using default value %d", key, defaultValue)
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("WARNING: Invalid boolean format for %s, using default value %v", key, defaultValue)
		return defaultValue
	}

	return value
}
