package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/middleware"
	"github.com/stretchr/testify/assert"
)

// TestAuthenticationFlow tests the complete authentication flow from registration to profile access
func TestAuthenticationFlow(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := SetupTestDB()
	cfg := GetTestConfig()

	// Create controllers
	authController := controllers.NewAuthController(db, cfg)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())

	// Register routes
	router.POST("/auth/register", authController.Register)
	router.POST("/auth/login", authController.Login)
	router.GET("/auth/profile", middleware.AuthMiddleware(cfg.JWTSecret), authController.GetCurrentUser)

	// Step 1: Register a new user
	t.Run("Register user", func(t *testing.T) {
		registerPayload := `{
			"name": "Integration Test User",
			"email": "integration-test@example.com",
			"password": "testpassword123"
		}`

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(registerPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse response to get token
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["token"])
	})

	// Step 2: Login with the registered user
	var authToken string
	t.Run("Login user", func(t *testing.T) {
		loginPayload := `{
			"email": "integration-test@example.com",
			"password": "testpassword123"
		}`

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response to get token
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["token"])

		// Save token for next step
		authToken = response["token"]
	})

	// Step 3: Access profile with token
	t.Run("Access profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Check user details
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Integration Test User", response["name"])
		assert.Equal(t, "integration-test@example.com", response["email"])
		assert.Equal(t, "student", response["role"])
	})
}
