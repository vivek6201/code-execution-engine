package worker

import (
	"context"

	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/pkg/cache"
	"github.com/code-execution-engine/pkg/db"
	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/judge/engine/evaluator"
	"github.com/code-execution-engine/internal/judge/engine/execution"
	"github.com/code-execution-engine/internal/judge/engine/executor"
	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/engine/runners/languages"
	"github.com/code-execution-engine/internal/judge/isolation"
	"github.com/code-execution-engine/pkg/queue"
	"github.com/code-execution-engine/pkg/telemetry"
	"github.com/code-execution-engine/internal/types"
	"github.com/google/uuid"
)

func StartWorker() {
	telemetry.Init()

	cfg := config.Load()
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient)

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
	service := execution.NewService(exec, eval)

	telemetry.Info("Worker started, listening for jobs")

	ctx := context.Background()
	for {
		jobID, j, err := q.Dequeue(ctx)
		if err != nil {
			telemetry.Error("Failed to dequeue job", "error", err)
			continue
		}

		loggerJobUUID, _ := uuid.Parse(jobID)

		// Mark processing in DB
		_ = judgeRepo.UpdateJobResult(loggerJobUUID, judge.UpdateJobDTO{
			Status: types.StatusProcessing,
		})

		telemetry.Info("Executing job", "job_id", jobID)
		res := service.Execute(j)

		err = queue.SetResult(ctx, redisClient, jobID, res)
		if err != nil {
			telemetry.Error("Failed to save result to redis", "job_id", jobID, "error", err)
		} else {
			telemetry.Info("Job finished in redis", "job_id", jobID)
		}

		// Update final state in Postgres
		dbErr := judgeRepo.UpdateJobResult(loggerJobUUID, judge.UpdateJobDTO{
			Status:     res.Status,
			Output:     res.Output,
			Error:      res.Error,
			FatalError: res.FatalError,
			TimeMs:     res.TimeMs,
			MemoryKB:   res.MemoryKB,
			Passed:     res.Passed,
			TestCases:  res.TestCases, // GORM json serializer handles this
		})

		if dbErr != nil {
			telemetry.Error("Failed to save result to postgres", "job_id", jobID, "error", dbErr)
		}
	}
}
