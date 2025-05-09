package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hesham-ashraf/LearnVibe/backend/cms/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test JWT secret
	jwtSecret := "test-secret"

	// Generate a valid token for testing
	testUserID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   testUserID.String(),
		"name":  "Test User",
		"email": "test@example.com",
		"role":  string(models.RoleStudent),
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	validTokenString, _ := token.SignedString([]byte(jwtSecret))

	// Generate an expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   testUserID.String(),
		"name":  "Test User",
		"email": "test@example.com",
		"role":  string(models.RoleStudent),
		"exp":   time.Now().Add(-time.Hour).Unix(), // Expired
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte(jwtSecret))

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantUserID bool
	}{
		{
			name:       "Valid Token",
			header:     "Bearer " + validTokenString,
			wantStatus: http.StatusOK,
			wantUserID: true,
		},
		{
			name:       "No Token",
			header:     "",
			wantStatus: http.StatusUnauthorized,
			wantUserID: false,
		},
		{
			name:       "Invalid Token Format",
			header:     "Bearer invalid-token",
			wantStatus: http.StatusUnauthorized,
			wantUserID: false,
		},
		{
			name:       "Expired Token",
			header:     "Bearer " + expiredTokenString,
			wantStatus: http.StatusUnauthorized,
			wantUserID: false,
		},
		{
			name:       "Missing Bearer Prefix",
			header:     validTokenString,
			wantStatus: http.StatusUnauthorized,
			wantUserID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router and protected route
			router := gin.New()
			router.GET("/protected", AuthMiddleware(jwtSecret), func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID not found"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"userID": userID})
			})

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, w.Code)

			// If we're expecting the middleware to pass
			if tt.wantStatus == http.StatusOK && tt.wantUserID {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, testUserID.String(), response["userID"])
			}
		})
	}
}

func TestRoleMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		role       models.Role
		middleware gin.HandlerFunc
		wantStatus int
	}{
		{
			name:       "Admin accessing admin route",
			role:       models.RoleAdmin,
			middleware: AdminOnly(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Instructor accessing admin route",
			role:       models.RoleInstructor,
			middleware: AdminOnly(),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Student accessing admin route",
			role:       models.RoleStudent,
			middleware: AdminOnly(),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Admin accessing instructor route",
			role:       models.RoleAdmin,
			middleware: InstructorOrAdmin(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Instructor accessing instructor route",
			role:       models.RoleInstructor,
			middleware: InstructorOrAdmin(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "Student accessing instructor route",
			role:       models.RoleStudent,
			middleware: InstructorOrAdmin(),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router with proper middleware
			router := gin.New()

			// Handler that sets role and processes middleware
			router.GET("/protected", func(c *gin.Context) {
				// Set userRole before middleware is applied - this matches the key used in auth.go
				c.Set("userRole", string(tt.role))

				// Call middleware directly
				tt.middleware(c)

				// If middleware didn't abort, send success
				if !c.IsAborted() {
					c.JSON(http.StatusOK, gin.H{"status": "success"})
				}
			})

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			w := httptest.NewRecorder()

			// Execute the request
			router.ServeHTTP(w, req)

			// Assert response status
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
