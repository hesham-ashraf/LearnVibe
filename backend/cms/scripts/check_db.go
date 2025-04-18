package main

import (
	"fmt"
	"log"

	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
)

func main() {
	fmt.Println("Connecting to database using environment variables...")

	// Initialize database with the new signature (no parameters)
	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Check if tables exist
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables)

	fmt.Println("\nTables in the database:")
	for i, table := range tables {
		fmt.Printf("%d. %s\n", i+1, table)
	}

	fmt.Println("\nDatabase connection and tables verified successfully!")
}
