package config

import (
	"os"

	"github.com/hesham-ashraf/LearnVibe/backend/gateway/config"
)

// TestConfig defines configuration for testing
type TestConfig struct {
	*config.Config
}

// LoadTestConfig loads a configuration for testing
func LoadTestConfig() (*TestConfig, error) {
	// Set testing environment variables
	os.Setenv("PORT", "8099") // Use different port for tests
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("CMS_SERVICE_URL", "http://localhost:8098") // Mock service URLs
	os.Setenv("CONTENT_SERVICE_URL", "http://localhost:8097")
	os.Setenv("ENABLE_HTTPS", "false") // Disable HTTPS for testing

	// Load configuration using the regular config loader
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &TestConfig{
		Config: cfg,
	}, nil
}

// ResetConfig restores environment variables to their original state
func ResetConfig() {
	// Unset all environment variables used in tests
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("CMS_SERVICE_URL")
	os.Unsetenv("CONTENT_SERVICE_URL")
	os.Unsetenv("ENABLE_HTTPS")
}
