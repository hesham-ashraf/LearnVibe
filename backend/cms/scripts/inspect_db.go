package main

import (
	"fmt"
	"log"

	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
)

func main() {
	// Initialize database connection
	db, err := models.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Inspect courses table
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		UdtName    string `gorm:"column:udt_name"`
	}

	result := db.Raw(`
		SELECT column_name, data_type, udt_name 
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'courses'
	`).Scan(&columns)

	if result.Error != nil {
		log.Fatal("Failed to query column information:", result.Error)
	}

	fmt.Println("COURSES TABLE COLUMNS:")
	fmt.Println("======================")
	for _, col := range columns {
		fmt.Printf("Column: %-15s | Data Type: %-15s | UDT Name: %s\n", col.ColumnName, col.DataType, col.UdtName)
	}
	fmt.Println()

	// Check if the course_contents table exists
	var tableExists int
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'course_contents'
	`).Scan(&tableExists)

	fmt.Printf("course_contents table exists: %v\n", tableExists > 0)
}
