package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/config"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// SimpleTestDB is a simple implementation of DBInterface for testing
type SimpleTestDB struct {
	users map[string]models.User
}

func NewSimpleTestDB() *SimpleTestDB {
	// Create a test user with password "password"
	testUser := models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleStudent,
	}
	testUser.SetPassword("password")

	// Create test db with user
	users := make(map[string]models.User)
	users[testUser.Email] = testUser

	return &SimpleTestDB{users: users}
}

func (db *SimpleTestDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	// We only support "email = ?" queries
	if queryStr, ok := query.(string); ok && queryStr == "email = ?" && len(args) > 0 {
		if email, ok := args[0].(string); ok {
			// Check if user exists
			if _, exists := db.users[email]; exists {
				return &gorm.DB{Error: nil}
			}
		}
	}
	return &gorm.DB{Error: nil}
}

func (db *SimpleTestDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	// Extract the condition value (email or id)
	if len(conds) > 0 {
		if condStr, ok := conds[0].(string); ok {
			if condStr == "id = ?" && len(conds) > 1 {
				// User lookup by ID
				if userID, ok := conds[1].(uuid.UUID); ok {
					// Find user by ID
					for _, user := range db.users {
						if user.ID == userID {
							if u, ok := dest.(*models.User); ok {
								*u = user
								return &gorm.DB{Error: nil}
							}
						}
					}
				}
				return &gorm.DB{Error: fmt.Errorf("user not found")}
			}
		}
	}

	// Otherwise, handle as if continuing from a Where clause
	if u, ok := dest.(*models.User); ok {
		// Just return the first user from our map
		for _, user := range db.users {
			*u = user
			return &gorm.DB{Error: nil}
		}
	}
	return &gorm.DB{Error: fmt.Errorf("user not found")}
}

func (db *SimpleTestDB) Create(value interface{}) *gorm.DB {
	if u, ok := value.(*models.User); ok {
		db.users[u.Email] = *u
		return &gorm.DB{Error: nil}
	}
	return &gorm.DB{Error: fmt.Errorf("invalid value type")}
}

// disabled_TestLogin tests the login functionality
func disabled_TestLogin(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testDB := NewSimpleTestDB()
	testConfig := &config.Config{
		JWTSecret: "test-secret",
	}

	// Create the controller with our test DB
	authController := NewAuthController(testDB, testConfig)

	// Create a new gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a request body
	loginRequest := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "password",
	}

	jsonValue, _ := json.Marshal(loginRequest)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call the function
	authController.Login(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response has token
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response["token"])
}

// disabled_TestRegister tests the user registration functionality
func disabled_TestRegister(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testDB := NewSimpleTestDB()
	testConfig := &config.Config{
		JWTSecret: "test-secret",
	}

	// Create the controller with our test DB
	authController := NewAuthController(testDB, testConfig)

	// Create a new gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a request body
	registerRequest := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
	}

	jsonValue, _ := json.Marshal(registerRequest)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call the function
	authController.Register(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify response has token
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response["token"])
}

// disabled_TestGetCurrentUser tests retrieval of the current user profile
func disabled_TestGetCurrentUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testDB := NewSimpleTestDB()
	testConfig := &config.Config{
		JWTSecret: "test-secret",
	}

	// Get test user ID (we'll use the one from our SimpleTestDB)
	var testUserID uuid.UUID
	for _, user := range testDB.users {
		testUserID = user.ID
		break
	}

	// Create a new gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user ID in gin context (as middleware would do)
	c.Set("userID", testUserID)

	// Create the controller with our test DB
	authController := NewAuthController(testDB, testConfig)

	// Call the function
	authController.GetCurrentUser(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user data in response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "Test User", response["name"])
}

// Additional tests would be added here for:
// - TestGoogleLogin
// - TestGoogleCallback
// - TestGenerateJWT
