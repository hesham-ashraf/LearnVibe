package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/services"
)

// RequestLoggerMiddleware logs HTTP requests using the centralized logging service
func RequestLoggerMiddleware(logger *services.LoggingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate request processing time
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		// Collect request details
		requestMethod := c.Request.Method
		requestURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Get user ID if authenticated
		userID, _ := c.Get("userID")

		// Create metadata for the log entry
		metadata := map[string]interface{}{
			"method":     requestMethod,
			"uri":        requestURI,
			"status":     statusCode,
			"latency_ms": latencyTime.Milliseconds(),
			"client_ip":  clientIP,
			"user_agent": userAgent,
		}

		// Add user ID if available
		if userID != nil {
			metadata["user_id"] = userID
		}

		// Determine log level based on status code
		if statusCode >= 500 {
			logger.Error("Server error in request processing", nil, metadata)
		} else if statusCode >= 400 {
			logger.Warning("Client error in request", metadata)
		} else {
			logger.Info("Request processed", metadata)
		}
	}
}
