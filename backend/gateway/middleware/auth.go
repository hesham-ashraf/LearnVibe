package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the structure for JWT token claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenValidationMiddleware validates JWT tokens and sets claims in context
func TokenValidationMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for auth endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/auth") ||
			strings.HasPrefix(c.Request.URL.Path, "/health") {
			c.Next()
			return
		}

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Check the Authorization header format
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check token expiration
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		// Set claims in context for future handlers
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)

		// Forward the authorization header to downstream services
		c.Request.Header.Set("X-User-ID", claims.UserID)
		c.Request.Header.Set("X-User-Role", claims.Role)

		c.Next()
	}
}
