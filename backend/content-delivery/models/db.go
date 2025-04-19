package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/sony/gobreaker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database represents our database connections
type Database struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// Load environment variables from .env file
func LoadEnv() {
	// We'll just attempt to load .env but not fail if it doesn't exist
	_ = godotenv.Load()
	// Log a message but continue execution
	log.Println("Loaded environment variables (if .env file exists)")
}

// InitDB initializes both SQL and Redis databases
func InitDB() (*Database, error) {
	// Load environment variables
	LoadEnv()

	// Initialize PostgreSQL
	db, err := initPostgres()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %v", err)
	}

	// Initialize Redis
	rdb, err := initRedis()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %v", err)
	}

	return &Database{
		DB:    db,
		Redis: rdb,
	}, nil
}

// initPostgres initializes PostgreSQL with retry and circuit breaker patterns
func initPostgres() (*gorm.DB, error) {
	// Get database URL from environment variables or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use a default connection string if environment variable is not set
		dbURL = "postgres://postgres:vampire8122003@localhost:5432/learnvibe_content"
		log.Println("DATABASE_URL not set, using default:", dbURL)
	} else {
		log.Println("Using DATABASE_URL from environment")
	}

	// Open a connection to the database with retry logic
	var db *gorm.DB
	var err error
	operation := func() error {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		return err
	}

	// Retry with exponential backoff for transient connection issues
	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Circuit Breaker setup for database
	settings := gobreaker.Settings{
		Name:    "DatabaseService",
		Timeout: 5 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	// Checking the database connection through the circuit breaker
	_, err = cb.Execute(func() (interface{}, error) {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}
		return sqlDB, nil
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	return db, nil
}

// initRedis initializes Redis with retry and circuit breaker patterns
func initRedis() (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
		log.Println("REDIS_URL not set, using default:", redisURL)
	}

	redisPass := os.Getenv("REDIS_PASSWORD")

	// Create Redis client
	var rdb *redis.Client
	operation := func() error {
		rdb = redis.NewClient(&redis.Options{
			Addr:        redisURL,
			Password:    redisPass,
			DB:          0, // default DB
			DialTimeout: 5 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Ping Redis to check the connection
		_, err := rdb.Ping(ctx).Result()
		return err
	}

	// Retry with exponential backoff for transient connection issues
	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to redis: %v", err)
	}

	// Circuit Breaker setup for Redis
	settings := gobreaker.Settings{
		Name:    "RedisService",
		Timeout: 5 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	// Checking the Redis connection through the circuit breaker
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cb.Execute(func() (interface{}, error) {
		return rdb.Ping(ctx).Result()
	})
	if err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	return rdb, nil
}

// MigrateDB performs database migrations with retry logic
func MigrateDB(db *gorm.DB) {
	// Migrate models in the correct order to avoid foreign key issues
	log.Println("Starting database migration...")

	log.Println("Migrating Content model...")
	if err := db.AutoMigrate(&Content{}); err != nil {
		log.Fatal("Error migrating Content model:", err)
	}

	log.Println("Database migration completed successfully!")
}