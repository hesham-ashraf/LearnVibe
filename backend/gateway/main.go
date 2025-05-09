package main

import (
	"log"
	"os"
	"path/filepath"

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

	// Create certs directory if it doesn't exist
	if cfg.EnableHTTPS {
		certsDir := "certs"
		if _, err := os.Stat(certsDir); os.IsNotExist(err) {
			err = os.MkdirAll(certsDir, 0755)
			if err != nil {
				log.Printf("Warning: Failed to create certs directory: %v", err)
			}
		}
	}

	// Start the server
	log.Printf("API Gateway running on port %s", cfg.Port)
	log.Printf("CMS Service URL: %s", cfg.CMSServiceURL)
	log.Printf("Content Service URL: %s", cfg.ContentServiceURL)

	// Start server with or without HTTPS
	if cfg.EnableHTTPS {
		certFile := filepath.Join("certs", "server.crt")
		keyFile := filepath.Join("certs", "server.key")

		// Check if certificate files exist
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.Printf("Warning: SSL certificate file not found at %s", certFile)
			log.Println("Falling back to HTTP")
			if err := router.Run(":" + cfg.Port); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			log.Printf("Warning: SSL key file not found at %s", keyFile)
			log.Println("Falling back to HTTP")
			if err := router.Run(":" + cfg.Port); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else {
			log.Printf("Starting server with HTTPS on port %s", cfg.HTTPSPort)
			if err := router.RunTLS(":"+cfg.HTTPSPort, certFile, keyFile); err != nil {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		}
	} else {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
