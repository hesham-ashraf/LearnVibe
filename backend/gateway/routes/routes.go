package routes

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/config"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/health"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/middleware"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/proxy"
)

// SetupRoutes configures all the routes for the API Gateway
func SetupRoutes(router *gin.Engine, serviceProxy *proxy.ServiceProxy, healthChecker *health.HealthChecker, cfg *config.Config) {
	// Middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitDuration))
	router.Use(middleware.TokenValidationMiddleware(cfg.JWTSecret))

	// Health check endpoint
	router.GET("/health", healthChecker.CheckAllServices)

	// Auth routes - proxy to CMS service
	router.Group("/auth/*path").Use(serviceProxy.ProxyCMSRequest())

	// CMS API routes
	cmsRoutes := []string{
		"/api/courses",
		"/api/enrollments",
		"/api/admin",
	}

	for _, route := range cmsRoutes {
		router.Group(route + "/*path").Use(serviceProxy.ProxyCMSRequest())
	}

	// Content API routes
	contentRoutes := []string{
		"/api/content",
		"/public/content",
	}

	for _, route := range contentRoutes {
		router.Group(route + "/*path").Use(serviceProxy.ProxyContentRequest())
	}

	// Fallback route - determine service based on path
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Decide where to proxy based on the path
		if strings.HasPrefix(path, "/api/content") || strings.HasPrefix(path, "/public/content") {
			serviceProxy.ProxyContentRequest()(c)
		} else {
			// Default to CMS service
			serviceProxy.ProxyCMSRequest()(c)
		}
	})
}
