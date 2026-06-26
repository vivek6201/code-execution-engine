package app

import (
	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/dashboard"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/server/middlewares"
	"github.com/code-execution-engine/internal/server/routes"
	"github.com/code-execution-engine/pkg/cache"
	"github.com/code-execution-engine/pkg/db"
	"github.com/code-execution-engine/pkg/queue"
	"github.com/code-execution-engine/pkg/telemetry"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	telemetry.Init()

	cfg := config.Load()

	// Infrastructure
	database := db.ConnectDB(cfg.DbUrl)
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient, cfg.RedisUrl)

	// Auth module
	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo, redisClient)
	authHandler := auth.NewHandler(authService)

	// Dashboard module
	dashRepo := dashboard.NewRepository(database)
	dashService := dashboard.NewService(dashRepo)
	dashHandler := dashboard.NewHandler(dashService)

	judgeRepo := judge.NewRepository(database)
	judgeHandler := judge.NewHandler(q, redisClient, judgeRepo)

	// Gin setup
	r := gin.Default()
	
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Session-ID"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	r.Use(middlewares.RequestLogger())

	routes.SetupRoutes(r, &routes.Dependencies{
		AuthHandler:  authHandler,
		AuthService:  authService,
		DashHandler:  dashHandler,
		JudgeHandler: judgeHandler,
		RedisClient:  redisClient,
	})

	telemetry.Info("API server starting", "port", 8080)
	r.Run(":8080")
}
