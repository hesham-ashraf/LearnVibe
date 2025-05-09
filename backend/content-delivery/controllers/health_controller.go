package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/models"
	"github.com/hesham-ashraf/LearnVibe/backend/content-delivery/services"
)

// HealthController handles health check endpoints
type HealthController struct {
	db            *models.Database
	messageBroker *services.MessageBroker
	logger        *services.LoggingService
}

// NewHealthController creates a new health controller
func NewHealthController(db *models.Database, messageBroker *services.MessageBroker, logger *services.LoggingService) *HealthController {
	return &HealthController{
		db:            db,
		messageBroker: messageBroker,
		logger:        logger,
	}
}

// CheckHealth handles the health check endpoint request
func (hc *HealthController) CheckHealth(c *gin.Context) {
	health := map[string]string{
		"status":   "UP",
		"service":  "content-delivery",
		"database": "UP",
		"redis":    "UP",
		"rabbitmq": "UP",
		"logging":  "UP",
	}

	// Check database connection
	sqlDB, err := hc.db.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		health["database"] = "DOWN"
		health["status"] = "DEGRADED"
	}

	// Check Redis connection
	ctx := c.Request.Context()
	_, err = hc.db.Redis.Ping(ctx).Result()
	if err != nil {
		health["redis"] = "DOWN"
		health["status"] = "DEGRADED"
	}

	// Check RabbitMQ if available
	if hc.messageBroker == nil {
		health["rabbitmq"] = "NOT_CONFIGURED"
		health["status"] = "DEGRADED"
	}

	// Response code depends on overall status
	responseCode := http.StatusOK
	if health["status"] != "UP" {
		responseCode = http.StatusServiceUnavailable
	}

	// Log health check results
	if hc.logger != nil {
		hc.logger.Info("Health check performed", map[string]interface{}{
			"health_status": health,
		})
	}

	c.JSON(responseCode, health)
}
