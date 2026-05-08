package middlewares

import (
	"net/http"
	"strings"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/code-execution-engine/internal/auth"
	"github.com/gin-gonic/gin"
)

// JWTAuth returns a Gin middleware that validates the Authorization: Bearer <token>
// header for dashboard routes. On success it sets "user_id" and "email" in
// the Gin context.
func JWTAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication failed", "invalid Authorization header format, expected 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(parts[1], authService.JWTSecret())
		if err != nil {
			utility.ErrorResponse(c, http.StatusUnauthorized, "Authentication failed", "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
