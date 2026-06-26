package worker

import (
	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/judge/engine/evaluator"
	"github.com/code-execution-engine/internal/judge/engine/execution"
	"github.com/code-execution-engine/internal/judge/engine/executor"
	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/engine/runners/languages"
	"github.com/code-execution-engine/internal/judge/isolation"
	"github.com/code-execution-engine/pkg/cache"
	"github.com/code-execution-engine/pkg/db"
	"github.com/code-execution-engine/pkg/queue"
	"github.com/code-execution-engine/pkg/telemetry"
	"github.com/hibiken/asynq"
)

func StartWorker() {
	telemetry.Init()

	cfg := config.Load()
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	database := db.ConnectDB(cfg.DbUrl)
	judgeRepo := judge.NewRepository(database)

	containerClient := isolation.NewClient()

	factory := runners.NewFactory()
	factory.Register("python", languages.NewPythonRunner(containerClient))
	factory.Register("javascript", languages.NewNodeRunner(containerClient))
	factory.Register("cpp", languages.NewCppRunner(containerClient))
	factory.Register("java", languages.NewJavaRunner(containerClient))

	exec := executor.NewExecutor(factory)
	eval := evaluator.NewService()
	executionService := execution.NewService(exec, eval)

	// Configure Asynq Server
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisUrl)
	if err != nil {
		redisOpt = asynq.RedisClientOpt{Addr: "localhost:6379"}
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	handler := NewHandler(redisClient, judgeRepo, executionService)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskRunCode, handler.HandleRunCodeTask)

	telemetry.Info("Worker started, listening for Asynq tasks")
	if err := srv.Run(mux); err != nil {
		telemetry.Error("Asynq server execution failed", "error", err)
		panic(err)
	}
}
