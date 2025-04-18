package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB initializes the database connection
func InitDB(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// MigrateDB performs database migrations
func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Course{}, &CourseContent{})
}
