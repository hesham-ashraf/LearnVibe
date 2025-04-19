package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/config"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/health"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/proxy"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/routes"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize service proxy
	serviceProxy, err := proxy.NewServiceProxy(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize service proxy: %v", err)
	}

	// Initialize health checker
	healthChecker := health.NewHealthChecker(cfg)

	// Setup Gin
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, serviceProxy, healthChecker, cfg)

	// Start the server
	log.Printf("API Gateway running on port %s", cfg.Port)
	log.Printf("CMS Service URL: %s", cfg.CMSServiceURL)
	log.Printf("Content Service URL: %s", cfg.ContentServiceURL)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
