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
func LoadEnv() {
	// We'll just attempt to load .env but not fail if it doesn't exist
	_ = godotenv.Load()
	// Log a message but continue execution
	log.Println("Loaded environment variables (if .env file exists)")
}

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	// Load environment variables
	LoadEnv()

	// Get database URL from environment variables or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use a default connection string if environment variable is not set
		dbURL = "postgres://postgres:vampire8122003@localhost:5432/learnvibe"
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
	// First drop the course_contents table if it exists to avoid constraint issues
	log.Println("Dropping CourseContent table if exists...")
	db.Exec("DROP TABLE IF EXISTS course_contents")

	log.Println("Creating CourseContent table manually...")
	// Create the course_contents table with explicit column types matching the existing schema
	query := `
	CREATE TABLE IF NOT EXISTS course_contents (
		id UUID PRIMARY KEY,
		course_id INTEGER NOT NULL,
		title TEXT,
		description TEXT,
		type VARCHAR(10),
		url TEXT,
		"order" INT,
		created_at TIMESTAMP WITH TIME ZONE,
		updated_at TIMESTAMP WITH TIME ZONE,
		CONSTRAINT fk_course_id FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
	)`
	if err := db.Exec(query).Error; err != nil {
		log.Fatal("Error creating CourseContent table:", err)
	}

	// Create index on course_id
	db.Exec("CREATE INDEX IF NOT EXISTS idx_course_contents_course_id ON course_contents(course_id)")

	log.Println("Migrating Enrollment model...")
	if err := db.AutoMigrate(&Enrollment{}); err != nil {
		log.Fatal("Error migrating Enrollment model:", err)
	}

	log.Println("Database migration completed successfully!")
}
