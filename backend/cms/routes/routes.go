package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/middleware"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine, courseController *controllers.CourseController,
	authController *controllers.AuthController, enrollmentController *controllers.EnrollmentController,
	healthController *controllers.HealthController, cfg *config.Config) {
	// Auth routes
	authRoutes := router.Group("/auth")
	{
		// OAuth routes
		authRoutes.GET("/google", authController.GoogleLogin)
		authRoutes.GET("/google/callback", authController.GoogleCallback)

		// Standard authentication routes
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)

		// Current user route (protected)
		authRoutes.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authController.GetCurrentUser)
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

			// Enrollment routes
			courses.POST("/:id/enroll", enrollmentController.EnrollInCourse)

			// Routes restricted to instructors and admins
			instructorRoutes := courses.Group("")
			instructorRoutes.Use(middleware.InstructorOrAdmin())
			{
				instructorRoutes.POST("", courseController.CreateCourse)
				instructorRoutes.PUT("/:id", courseController.UpdateCourse)
				instructorRoutes.DELETE("/:id", courseController.DeleteCourse)
				instructorRoutes.POST("/:id/contents", courseController.AddCourseContent)
				instructorRoutes.DELETE("/:id/contents/:contentId", courseController.DeleteCourseContent)

				// View enrollments for a course (instructors/admins only)
				instructorRoutes.GET("/:id/enrollments", enrollmentController.GetCourseEnrollments)
			}
		}

		// Enrollment management routes
		enrollments := api.Group("/enrollments")
		{
			// User enrollment management
			enrollments.GET("", enrollmentController.GetUserEnrollments)
			enrollments.GET("/:id", enrollmentController.GetEnrollmentDetails)
			enrollments.PUT("/:id/progress", enrollmentController.UpdateEnrollmentProgress)
			enrollments.PUT("/:id/drop", enrollmentController.DropEnrollment)
		}

		// Admin-only routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			// TODO: Add admin-specific routes here if needed
		}
	}

	// Health check
	router.GET("/health", healthController.CheckHealth)
}
