package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/gateway/config"
)

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// GatewayStatus represents the status of the API gateway and its services
type GatewayStatus struct {
	Status   string          `json:"status"`
	Services []ServiceStatus `json:"services"`
	Time     time.Time       `json:"time"`
}

// HealthChecker performs health checks on microservices
type HealthChecker struct {
	cmsURL        string
	contentURL    string
	cmsHealth     string
	contentHealth string
	client        *http.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.Config) *HealthChecker {
	return &HealthChecker{
		cmsURL:        cfg.CMSServiceURL,
		contentURL:    cfg.ContentServiceURL,
		cmsHealth:     cfg.CMSHealthEndpoint,
		contentHealth: cfg.ContentHealthEndpoint,
		client: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
	}
}

// checkServiceHealth checks the health of a service
func (hc *HealthChecker) checkServiceHealth(ctx context.Context, baseURL, healthEndpoint, serviceName string) ServiceStatus {
	url := fmt.Sprintf("%s%s", baseURL, healthEndpoint)
	status := ServiceStatus{
		Name: serviceName,
		URL:  baseURL,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		status.Status = "error"
		return status
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		status.Status = "down"
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status.Status = "up"
	} else {
		status.Status = "down"
	}

	return status
}

// CheckAllServices checks the health of all services
func (hc *HealthChecker) CheckAllServices(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check all services concurrently
	var wg sync.WaitGroup
	serviceStatuses := make([]ServiceStatus, 2)

	wg.Add(2)

	// Check CMS service
	go func() {
		defer wg.Done()
		serviceStatuses[0] = hc.checkServiceHealth(ctx, hc.cmsURL, hc.cmsHealth, "CMS Service")
	}()

	// Check Content service
	go func() {
		defer wg.Done()
		serviceStatuses[1] = hc.checkServiceHealth(ctx, hc.contentURL, hc.contentHealth, "Content Service")
	}()

	wg.Wait()

	// Determine overall status
	overallStatus := "up"
	for _, status := range serviceStatuses {
		if status.Status != "up" {
			overallStatus = "degraded"
			if status.Status == "down" {
				// If any critical service is down, mark as down
				if status.Name == "CMS Service" {
					overallStatus = "down"
					break
				}
			}
		}
	}

	// Return the health check results
	c.JSON(http.StatusOK, GatewayStatus{
		Status:   overallStatus,
		Services: serviceStatuses,
		Time:     time.Now(),
	})
}
