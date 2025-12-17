package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(client *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Identify user by IP (or UserID if logged in)
		identifier := c.ClientIP()

		key := fmt.Sprintf("rate_limit:%s", identifier)

		// Increment count
		count, err := client.Incr(c.Request.Context(), key).Result()
		if err != nil {
			c.Next() // Fail open (allow request) if Redis errors
			return
		}

		// Set expiration on first request
		if count == 1 {
			client.Expire(c.Request.Context(), key, 1*time.Minute)
		}

		// Check limit (e.g., 20 requests per minute)
		if count > 20 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
