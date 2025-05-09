package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/controllers"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/routes"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB       *gorm.DB
	testRouter   *gin.Engine
	testConfig   *config.Config
	jwtToken     string
	testUserID   string
	testUserName = "Integration Test User"
	testEmail    = "integration-test@example.com"
	testPassword = "Password123!"
)

// setupTestDB initializes a test database connection
func setupTestDB() (*gorm.DB, error) {
	// This would ideally be a test database or container
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/learnvibe_test"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the test database
	err = db.AutoMigrate(&models.User{}, &models.Course{}, &models.Enrollment{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// setupTestRouter creates a router with all routes for integration testing
func setupTestRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Initialize controllers
	courseController := controllers.NewCourseController(db)
	authController := controllers.NewAuthController(db, cfg)
	enrollmentController := controllers.NewEnrollmentController(db)
	healthController := controllers.NewTestHealthController()

	// Setup routes
	routes.SetupRoutes(router, courseController, authController, enrollmentController, healthController, cfg)

	return router
}

// TestMain sets up the testing environment
func TestMain(m *testing.M) {
	var err error

	// Load test configuration
	testConfig = &config.Config{
		JWTSecret: "integration-test-secret",
	}

	// Setup test database
	testDB, err = setupTestDB()
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Setup router
	testRouter = setupTestRouter(testDB, testConfig)

	// Run the tests
	exitCode := m.Run()

	// Cleanup (would drop test database or tables here)
	sqlDB, _ := testDB.DB()
	sqlDB.Close()

	os.Exit(exitCode)
}

// TestAuthFlow tests the complete auth flow from registration to profile access
func TestAuthFlow(t *testing.T) {
	// 1. Test Registration
	t.Run("Register", func(t *testing.T) {
		// Create a registration request
		registerPayload := map[string]interface{}{
			"name":     testUserName,
			"email":    testEmail,
			"password": testPassword,
		}

		jsonValue, _ := json.Marshal(registerPayload)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert successful registration
		assert.Equal(t, http.StatusCreated, resp.Code)

		// Extract and store token
		var result map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result["token"])
		jwtToken = result["token"]
	})

	// 2. Test Login
	t.Run("Login", func(t *testing.T) {
		// Create a login request
		loginPayload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		jsonValue, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert successful login
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify token
		var result map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result["token"])
	})

	// 3. Test Get Current User
	t.Run("GetCurrentUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert successful response
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify user data
		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, testEmail, result["email"])
		assert.Equal(t, testUserName, result["name"])

		// Store user ID for other tests
		testUserID = result["id"].(string)
	})
}

// TestInvalidAuth tests authentication failures
func TestInvalidAuth(t *testing.T) {
	// Test login with invalid credentials
	t.Run("InvalidLogin", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    testEmail,
			"password": "wrong-password",
		}

		jsonValue, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert unauthorized status
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// Test accessing protected route without auth
	t.Run("AccessWithoutAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert unauthorized status
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	// Test accessing protected route with invalid token
	t.Run("AccessWithInvalidToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp := httptest.NewRecorder()
		testRouter.ServeHTTP(resp, req)

		// Assert unauthorized status
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
