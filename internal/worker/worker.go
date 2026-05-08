package worker

import (
	"context"

	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internal/cache"
	"github.com/code-execution-engine/internal/judge/engine/evaluator"
	"github.com/code-execution-engine/internal/judge/engine/execution"
	"github.com/code-execution-engine/internal/judge/engine/executor"
	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/engine/runners/languages"
	"github.com/code-execution-engine/internal/judge/isolation"
	"github.com/code-execution-engine/internal/queue"
	"github.com/code-execution-engine/internal/telemetry"
)

func StartWorker() {
	telemetry.Init()

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

	telemetry.Info("Worker started, listening for jobs")

	ctx := context.Background()
	for {
		jobID, j, err := q.Dequeue(ctx)
		if err != nil {
			telemetry.Error("Failed to dequeue job", "error", err)
			continue
		}

		telemetry.Info("Executing job", "job_id", jobID)
		res := service.Execute(j)

		err = queue.SetResult(ctx, redisClient, jobID, res)
		if err != nil {
			telemetry.Error("Failed to save result", "job_id", jobID, "error", err)
		} else {
			telemetry.Info("Job finished", "job_id", jobID)
		}
	}
}
