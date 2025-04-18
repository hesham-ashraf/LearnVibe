package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/middleware"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, courseController *controllers.CourseController, authController *controllers.AuthController, cfg *config.Config) {
	// Auth routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.GET("/google", authController.GoogleLogin)
		authRoutes.GET("/google/callback", authController.GoogleCallback)
	}

	// API routes (protected)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Course routes
		courses := api.Group("/courses")
		{
			// Routes accessible to all authenticated users
			courses.GET("", courseController.GetCourses)
			courses.GET("/:id", courseController.GetCourse)

			// Routes restricted to instructors and admins
			instructorRoutes := courses.Group("")
			instructorRoutes.Use(middleware.InstructorOrAdmin())
			{
				instructorRoutes.POST("", courseController.CreateCourse)
				instructorRoutes.PUT("/:id", courseController.UpdateCourse)
				instructorRoutes.DELETE("/:id", courseController.DeleteCourse)
				instructorRoutes.POST("/:id/contents", courseController.AddCourseContent)
				instructorRoutes.DELETE("/:id/contents/:contentId", courseController.DeleteCourseContent)
			}
		}

		// Admin-only routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			// TODO: Add admin-specific routes here if needed
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}
