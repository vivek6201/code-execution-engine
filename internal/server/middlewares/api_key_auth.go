package middlewares

import (
	"net/http"
	"strings"

	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/models"
	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth returns a Gin middleware that validates the X-API-Key header
// against the auth service. On success it sets "user" and "api_key" in the
// Gin context for downstream handlers.
func APIKeyAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "missing X-API-Key header")
			c.Abort()
			return
		}

		// Basic format check
		if !strings.HasPrefix(rawKey, "cee_") {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication failed", "invalid API key format")
			c.Abort()
			return
		}

		user, apiKey, err := authService.ValidateAPIKey(rawKey)
		if err != nil {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication failed", "invalid or revoked API key")
			c.Abort()
			return
		}

		// Make user and key info available to handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("api_key", apiKey)
		c.Set("plan_limits", models.Plans[user.Plan])

		c.Next()
	}
}
