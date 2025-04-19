package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the structure for JWT token claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware verifies the JWT token and sets user info in the context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check the Authorization header format
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := headerParts[1]

		// Parse and validate the token
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		// Handle token parsing errors
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}

		// Set user ID and role in context for future handlers
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// InstructorOrAdmin middleware ensures the user is an instructor or admin
func InstructorOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user has instructor or admin role
		userRole := role.(string)
		if userRole != "instructor" && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "This operation requires instructor or admin privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly middleware ensures the user is an admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user has admin role
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "This operation requires admin privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}
