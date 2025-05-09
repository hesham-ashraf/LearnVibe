package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/services"
	"gorm.io/gorm"
)

// HealthController handles health check requests
type HealthController struct {
	db            *gorm.DB
	messageBroker *services.MessageBroker
	logger        *services.LoggingService
}

// NewHealthController creates a new health controller
func NewHealthController(db *gorm.DB, messageBroker *services.MessageBroker, logger *services.LoggingService) *HealthController {
	return &HealthController{
		db:            db,
		messageBroker: messageBroker,
		logger:        logger,
	}
}

// NewTestHealthController creates a simplified health controller for testing
func NewTestHealthController() *HealthController {
	return &HealthController{
		db:            nil,
		messageBroker: nil,
		logger:        nil,
	}
}

// CheckHealth checks the health of the service and its dependencies
func (hc *HealthController) CheckHealth(c *gin.Context) {
	status := http.StatusOK
	response := gin.H{
		"status":  "ok",
		"service": "cms",
	}

	// Check database connection if provided
	if hc.db != nil {
		sqlDB, err := hc.db.DB()
		if err != nil {
			response["database"] = "error"
			response["database_error"] = err.Error()
			status = http.StatusServiceUnavailable
		} else if err := sqlDB.Ping(); err != nil {
			response["database"] = "error"
			response["database_error"] = err.Error()
			status = http.StatusServiceUnavailable
		} else {
			response["database"] = "ok"
		}
	}

	// Check message broker if provided
	if hc.messageBroker != nil {
		// Simplified check - in a real scenario you might want to check if connections are alive
		response["message_broker"] = "ok"
	}

	c.JSON(status, response)
}
