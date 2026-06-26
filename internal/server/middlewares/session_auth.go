package middlewares

import (
	"net/http"
	"strings"

	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
)

// SessionAuth returns a Gin middleware that validates user sessions.
// It checks cookies first, then Authorization bearer token, then the X-Session-ID header.
func SessionAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionID string

		// 1. Try Cookie
		cookieVal, err := c.Cookie("session_id")
		if err == nil && cookieVal != "" {
			sessionID = cookieVal
		}

		// 2. Try Authorization Header fallback
		if sessionID == "" {
			header := c.GetHeader("Authorization")
			if header != "" {
				parts := strings.SplitN(header, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					sessionID = parts[1]
				}
			}
		}

		// 3. Try X-Session-ID Header fallback
		if sessionID == "" {
			sessionID = c.GetHeader("X-Session-ID")
		}

		// If no session ID found, abort
		if sessionID == "" {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "missing session token")
			c.Abort()
			return
		}

		// Validate against Redis session store
		sessionData, err := authService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication failed", err.Error())
			c.Abort()
			return
		}

		// Inject details into Gin request context
		c.Set("user_id", sessionData.UserID)
		c.Set("email", sessionData.Email)
		c.Set("session_id", sessionID)

		c.Next()
	}
}
