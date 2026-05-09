package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/code-execution-engine/internal/models"
	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RateLimiter returns a Gin middleware that enforces per-user daily request
// limits based on the user's subscription plan. Uses a Redis key with a
// 24-hour TTL for each user+date combination.
//
// Must be used AFTER APIKeyAuth (requires "user_id" and "plan" in context).
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "user not authenticated")
			c.Abort()
			return
		}

		plan, _ := c.Get("plan")
		planType, ok := plan.(models.PlanType)
		if !ok {
			planType = models.PlanBasic
		}

		limits, exists := models.Plans[planType]
		if !exists {
			limits = models.Plans[models.PlanBasic]
		}

		// Build a daily rate limit key: ratelimit:{user_id}:{YYYY-MM-DD}
		date := time.Now().UTC().Format("2006-01-02")
		key := fmt.Sprintf("ratelimit:%s:%s", userID.(uuid.UUID).String(), date)

		ctx := c.Request.Context()

		// Increment and get current count
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow the request (fail open)
			c.Next()
			return
		}

		// Set TTL on first increment
		if count == 1 {
			redisClient.Expire(ctx, key, 24*time.Hour)
		}

		// Set rate limit headers
		remaining := limits.RequestsPerDay - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limits.RequestsPerDay))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if int(count) > limits.RequestsPerDay {
			c.Header("Retry-After", "86400") // 24 hours in seconds
			utility.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded",
				fmt.Sprintf("daily limit of %d requests exceeded for %s plan", limits.RequestsPerDay, planType))
			c.Abort()
			return
		}

		c.Next()
	}
}
