package main

import (
	"log"

	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/middleware"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := models.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the models
	models.MigrateDB(db)

	// Initialize controllers
	courseController := controllers.NewCourseController(db)
	authController := controllers.NewAuthController(db, cfg)

	// Initialize router
	router := gin.Default()

	// Setup middleware
	router.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(router, courseController, authController, cfg)

	// Start the server
	log.Printf("Server running on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
