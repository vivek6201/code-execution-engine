package app

import (
	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/cache"
	"github.com/code-execution-engine/internal/db"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/queue"
	"github.com/code-execution-engine/internal/server/middlewares"
	"github.com/code-execution-engine/internal/server/routes"
	"github.com/code-execution-engine/internal/telemetry"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	telemetry.Init()

	cfg := config.Load()

	// Infrastructure
	database := db.ConnectDB(cfg.DbUrl)
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient)

	// Auth module
	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	// Judge module
	judgeHandler := judge.NewHandler(q, redisClient)

	// Gin setup
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(middlewares.RequestLogger())

	routes.SetupRoutes(r, &routes.Dependencies{
		AuthHandler:  authHandler,
		AuthService:  authService,
		JudgeHandler: judgeHandler,
		RedisClient:  redisClient,
	})

	telemetry.Info("API server starting", "port", 8080)
	r.Run(":8080")
}
