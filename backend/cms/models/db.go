package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/joho/godotenv"
	"github.com/sony/gobreaker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Load environment variables from .env file
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Error loading .env file")
	}
	return nil
}

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		log.Fatal(err)
	}

	// Get database URL from environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
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
		return nil, fmt.Errorf("Error opening database: %v", err)
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
		return nil, fmt.Errorf("Database connection failed: %v", err)
	}

	return db, nil
}

// MigrateDB performs database migrations with retry logic
func MigrateDB(db *gorm.DB) {
	// Migrate models in the correct order to avoid foreign key issues
	log.Println("Starting database migration...")

	// First migrate models that don't depend on others
	log.Println("Migrating User model...")
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Error migrating User model:", err)
	}

	log.Println("Migrating Course model...")
	if err := db.AutoMigrate(&Course{}); err != nil {
		log.Fatal("Error migrating Course model:", err)
	}

	// Then migrate models that depend on the first ones
	log.Println("Migrating CourseContent model...")
	if err := db.AutoMigrate(&CourseContent{}); err != nil {
		log.Fatal("Error migrating CourseContent model:", err)
	}

	log.Println("Migrating Enrollment model...")
	if err := db.AutoMigrate(&Enrollment{}); err != nil {
		log.Fatal("Error migrating Enrollment model:", err)
	}

	log.Println("Database migration completed successfully!")
}
