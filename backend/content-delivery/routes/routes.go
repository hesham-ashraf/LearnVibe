package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/config"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/middleware"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, contentController *controllers.ContentController, healthController *controllers.HealthController, cfg *config.Config) {
	// Health check
	router.GET("/health", healthController.CheckHealth)

	// API routes (protected)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Content routes
		content := api.Group("/content")
		{
			// Routes accessible to all authenticated users
			content.GET("/:id", contentController.GetContent)
			content.GET("/:id/download", contentController.GetContentDownloadURL)

			// Routes restricted to instructors and admins
			instructorRoutes := content.Group("")
			instructorRoutes.Use(middleware.InstructorOrAdmin())
			{
				instructorRoutes.POST("", contentController.UploadContent)
				instructorRoutes.DELETE("/:id", contentController.DeleteContent)
			}
		}

		// For direct public access to content without authentication
		// This would typically be used for publicly available content
		// or for content that's served through signed URLs
		public := router.Group("/public")
		{
			// These routes would be implemented in a real system
			// They would likely validate signed URLs to control access
			public.GET("/content/:id", contentController.GetContent)
		}
	}
}
