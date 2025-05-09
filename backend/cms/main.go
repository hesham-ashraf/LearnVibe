package main

import (
	"log"

	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/middleware"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/routes"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize centralized logging service
	logger, err := services.NewLoggingService("cms-service", cfg.OpenSearchURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize OpenSearch logging service: %v", err)
		log.Println("Continuing with local logging only...")
	} else {
		defer logger.Close()
	}

	// Initialize message broker (RabbitMQ)
	messageBroker, err := services.NewMessageBroker(cfg.RabbitMQURL, cfg.RabbitMQExchange)
	if err != nil {
		logger.Warning("Failed to initialize RabbitMQ message broker. Some features may be limited.", nil)
		log.Printf("Warning: Failed to initialize RabbitMQ: %v", err)
	} else {
		defer messageBroker.Close()
		logger.Info("Successfully connected to RabbitMQ message broker", nil)
	}

	// Initialize database
	db, err := models.InitDB()
	if err != nil {
		logger.Fatal("Failed to connect to database", err, nil)
	}
	logger.Info("Successfully connected to the database", map[string]interface{}{
		"database": "PostgreSQL",
	})

	// Auto-migrate the models
	models.MigrateDB(db)
	logger.Info("Database migration completed", nil)

	// Initialize controllers with required services
	courseController := controllers.NewCourseController(db)
	authController := controllers.NewAuthController(db, cfg)
	enrollmentController := controllers.NewEnrollmentController(db)

	// Set up a health check handler that also monitors RabbitMQ and OpenSearch
	healthController := controllers.NewHealthController(db, messageBroker, logger)

	// Initialize router
	router := gin.Default()

	// Setup middleware
	router.Use(middleware.CORSMiddleware())
	// Add a request logging middleware
	router.Use(middleware.RequestLoggerMiddleware(logger))

	// Setup routes
	routes.SetupRoutes(router, courseController, authController, enrollmentController, healthController, cfg)

	// Start the server
	logger.Info("Server starting", map[string]interface{}{
		"port": cfg.Port,
	})
	log.Printf("Server running on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", err, nil)
	}
}
