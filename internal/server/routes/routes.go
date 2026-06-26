package routes

import (
	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/dashboard"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Dependencies holds all the handler and service instances needed by the router.
type Dependencies struct {
	AuthHandler  *auth.Handler
	AuthService  *auth.Service
	DashHandler  *dashboard.Handler
	JudgeHandler *judge.Handler
	RedisClient  *redis.Client
}

// SetupRoutes registers all API route groups.
func SetupRoutes(r *gin.Engine, deps *Dependencies) {
	apiv1 := r.Group("/api/v1")

	// Public metadata routes
	apiv1.GET("/languages", judge.GetLanguages)
	apiv1.GET("/statuses", judge.GetStatuses)

	// Public auth routes (no middleware)
	pub := apiv1.Group("/auth")
	{
		pub.POST("/register", deps.AuthHandler.Register)
		pub.POST("/login", deps.AuthHandler.Login)
	}

	// Dashboard routes (Session protected)
	dash := apiv1.Group("/dashboard")
	dash.Use(middlewares.SessionAuth(deps.AuthService))
	{
		dash.GET("/me", deps.DashHandler.GetMe)
		dash.PATCH("/plan", deps.DashHandler.UpdatePlan)
		dash.POST("/keys", deps.DashHandler.CreateKey)
		dash.GET("/keys", deps.DashHandler.ListKeys)
		dash.DELETE("/keys/:id", deps.DashHandler.RevokeKey)
		dash.GET("/jobs", deps.DashHandler.GetJobs)
		dash.GET("/metrics", deps.DashHandler.GetMetrics)
		dash.POST("/logout", deps.AuthHandler.Logout)
	}

	// Judge API routes (API key auth + rate limiting)
	j := apiv1.Group("/judge")
	j.Use(middlewares.APIKeyAuth(deps.AuthService))
	j.Use(middlewares.RateLimiter(deps.RedisClient))
	{
		j.POST("/run", deps.JudgeHandler.RunCode)
		j.GET("/run/:id", deps.JudgeHandler.GetResult)
	}
}
