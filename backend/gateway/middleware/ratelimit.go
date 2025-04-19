package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters for different client IPs
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// GetLimiter gets or creates a rate limiter for a client
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// RateLimitMiddleware applies rate limiting based on client IP
func RateLimitMiddleware(requestsPerSecond, burst int) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(requestsPerSecond, burst)

	return func(c *gin.Context) {
		// Skip rate limiting for health check endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Get client IP
		ip := c.ClientIP()
		limiter := rateLimiter.GetLimiter(ip)

		// Check if rate limit is exceeded
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
