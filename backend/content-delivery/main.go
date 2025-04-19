package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/config"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/middleware"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/models"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/routes"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connections
	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the models
	models.MigrateDB(db.DB)

	// Initialize storage service
	storageService, err := services.NewStorageService(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	// Initialize controllers
	contentController := controllers.NewContentController(db, storageService)

	// Initialize router
	router := gin.Default()

	// Setup middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(router, contentController, cfg)

	// Start the server
	log.Printf("Content Delivery Service running on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}