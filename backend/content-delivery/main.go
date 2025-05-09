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

	// Initialize centralized logging service
	logger, err := services.NewLoggingService("content-delivery-service", cfg.OpenSearchURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize OpenSearch logging service: %v", err)
		log.Println("Continuing with local logging only...")
	} else {
		defer logger.Close()
	}

	// Initialize message broker (RabbitMQ)
	messageBroker, err := services.NewMessageBroker(cfg.RabbitMQURL, cfg.RabbitMQExchange)
	if err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ: %v", err)
		if logger != nil {
			logger.Warning("Failed to initialize RabbitMQ message broker. Some features may be limited.", nil)
		}
	} else {
		defer messageBroker.Close()
		if logger != nil {
			logger.Info("Successfully connected to RabbitMQ message broker", nil)
		}
	}

	// Initialize database connections
	db, err := models.InitDB()
	if err != nil {
		if logger != nil {
			logger.Fatal("Failed to connect to database", err, nil)
		} else {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	}

	if logger != nil {
		logger.Info("Successfully connected to the database and Redis", nil)
	}

	// Auto-migrate the models
	models.MigrateDB(db.DB)
	if logger != nil {
		logger.Info("Database migration completed", nil)
	}

	// Initialize storage service
	storageService, err := services.NewStorageService(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		if logger != nil {
			logger.Fatal("Failed to initialize storage service", err, nil)
		} else {
			log.Fatalf("Failed to initialize storage service: %v", err)
		}
	}
	if logger != nil {
		logger.Info("Successfully connected to MinIO storage service", nil)
	}

	// Initialize controllers
	contentController := controllers.NewContentController(db, storageService)
	healthController := controllers.NewHealthController(db, messageBroker, logger)

	// Initialize router
	router := gin.Default()

	// Setup middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	if logger != nil {
		router.Use(middleware.RequestLoggerMiddleware(logger))
	}

	// Setup routes
	routes.SetupRoutes(router, contentController, healthController, cfg)

	// Start the server
	if logger != nil {
		logger.Info("Content Delivery Service starting", map[string]interface{}{
			"port": cfg.Port,
		})
	}
	log.Printf("Content Delivery Service running on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		if logger != nil {
			logger.Fatal("Failed to start server", err, nil)
		} else {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
