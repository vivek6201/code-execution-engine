package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	"github.com/code-execution-engine/internal/types"
	"github.com/google/uuid"
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

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskRunCode, func(ctx context.Context, t *asynq.Task) error {
		var payload queue.RunCodePayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			telemetry.Error("Failed to unmarshal task payload", "error", err)
			return err
		}

		jobUUID, err := uuid.Parse(payload.JobID)
		if err != nil {
			telemetry.Error("Invalid job UUID in task", "job_id", payload.JobID, "error", err)
			return err
		}

		// Update database status to PROCESSING
		_ = judgeRepo.UpdateJobResult(jobUUID, judge.UpdateJobDTO{
			Status: types.StatusProcessing,
		})
		_ = queue.SetStatus(ctx, redisClient, payload.JobID, types.StatusProcessing)

		telemetry.Info("Executing queued job", "job_id", payload.JobID)

		// Defaults for limits if somehow omitted in payload
		var timeLimit int64 = 2000
		if payload.TimeLimitMS != nil {
			timeLimit = *payload.TimeLimitMS
		}
		var memLimit int64 = 128000
		if payload.MemoryLimitKB != nil {
			memLimit = *payload.MemoryLimitKB
		}

		j := types.Job{
			Code:          payload.Code,
			Language:      payload.Language,
			Input:         payload.Input,
			TestCases:     payload.TestCases,
			TimeLimitMS:   timeLimit,
			MemoryLimitKB: memLimit,
		}

		res := executionService.Execute(j)

		// Cache final result in Redis (polling optimization)
		err = queue.SetResult(ctx, redisClient, payload.JobID, res)
		if err != nil {
			telemetry.Error("Failed to save result to Redis", "job_id", payload.JobID, "error", err)
		} else {
			telemetry.Info("Saved execution result to Redis", "job_id", payload.JobID)
		}

		// Save final result in Postgres
		dbErr := judgeRepo.UpdateJobResult(jobUUID, judge.UpdateJobDTO{
			Status:     res.Status,
			Output:     res.Output,
			Error:      res.Error,
			FatalError: res.FatalError,
			TimeMs:     res.TimeMs,
			MemoryKB:   res.MemoryKB,
			Passed:     res.Passed,
			TestCases:  res.TestCases,
		})
		if dbErr != nil {
			telemetry.Error("Failed to save result to Postgres", "job_id", payload.JobID, "error", dbErr)
		}

		// Dispatch webhook callback if callback_url is registered
		if payload.CallbackURL != nil && *payload.CallbackURL != "" {
			go triggerWebhookCallback(*payload.CallbackURL, payload.JobID, res)
		}

		return nil
	})

	telemetry.Info("Worker started, listening for Asynq tasks")
	if err := srv.Run(mux); err != nil {
		telemetry.Error("Asynq server execution failed", "error", err)
		panic(err)
	}
}

// Webhook payload schema
type webhookPayload struct {
	JobID      string                 `json:"job_id"`
	Status     types.Status           `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	FatalError string                 `json:"fatal_error,omitempty"`
	TimeMs     int64                  `json:"time_ms,omitempty"`
	MemoryKB   int64                  `json:"memory_kb,omitempty"`
	Passed     int                    `json:"passed,omitempty"`
	Total      int                    `json:"total,omitempty"`
	TestCases  []types.TestCaseResult `json:"test_cases,omitempty"`
}

func triggerWebhookCallback(callbackURL, jobID string, res types.Result) {
	payload := webhookPayload{
		JobID:      jobID,
		Status:     res.Status,
		Output:     res.Output,
		Error:      res.Error,
		FatalError: res.FatalError,
		TimeMs:     res.TimeMs,
		MemoryKB:   res.MemoryKB,
		Passed:     res.Passed,
		Total:      res.Total,
		TestCases:  res.TestCases,
	}

	bytesPayload, err := json.Marshal(payload)
	if err != nil {
		telemetry.Error("Failed to marshal webhook payload", "job_id", jobID, "error", err)
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(bytesPayload))
	if err != nil {
		telemetry.Error("Failed to create webhook request", "job_id", jobID, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Code-Execution-Engine-Webhook/1.0")

	resp, err := client.Do(req)
	if err != nil {
		telemetry.Error("Webhook callback delivery failed", "job_id", jobID, "url", callbackURL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		telemetry.Error("Webhook callback responded with non-2xx status code", "job_id", jobID, "url", callbackURL, "status", resp.Status)
	} else {
		telemetry.Info("Webhook callback delivered successfully", "job_id", jobID, "url", callbackURL, "status", resp.Status)
	}
}
