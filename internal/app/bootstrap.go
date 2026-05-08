package app

import (
	"context"

	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internal/auth"
	"github.com/code-execution-engine/internal/cache"
	"github.com/code-execution-engine/internal/db"
	"github.com/code-execution-engine/internal/engine/evaluator"
	"github.com/code-execution-engine/internal/engine/execution"
	"github.com/code-execution-engine/internal/engine/executor"
	"github.com/code-execution-engine/internal/engine/runners"
	"github.com/code-execution-engine/internal/engine/runners/languages"
	"github.com/code-execution-engine/internal/isolation"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/logger"
	"github.com/code-execution-engine/internal/queue"
	"github.com/code-execution-engine/internal/server/middlewares"
	"github.com/code-execution-engine/internal/server/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	logger.Init()

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

	logger.Info("API server starting", "port", 8080)
	r.Run(":8080")
}

func StartWorker() {
	logger.Init()

	cfg := config.Load()
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient)

	containerClient := isolation.NewClient()

	factory := runners.NewFactory()
	factory.Register("python", languages.NewPythonRunner(containerClient))
	factory.Register("javascript", languages.NewNodeRunner(containerClient))
	factory.Register("cpp", languages.NewCppRunner(containerClient))
	factory.Register("java", languages.NewJavaRunner(containerClient))

	exec := executor.NewExecutor(factory)
	eval := evaluator.NewService()
	service := execution.NewService(exec, eval)

	logger.Info("Worker started, listening for jobs")

	ctx := context.Background()
	for {
		jobID, j, err := q.Dequeue(ctx)
		if err != nil {
			logger.Error("Failed to dequeue job", "error", err)
			continue
		}

		logger.Info("Executing job", "job_id", jobID)
		res := service.Execute(j)

		err = queue.SetResult(ctx, redisClient, jobID, res)
		if err != nil {
			logger.Error("Failed to save result", "job_id", jobID, "error", err)
		} else {
			logger.Info("Job finished", "job_id", jobID)
		}
	}
}
