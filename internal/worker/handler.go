package worker

import (
	"context"
	"encoding/json"

	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/judge/engine/execution"
	"github.com/code-execution-engine/internal/types"
	"github.com/code-execution-engine/pkg/queue"
	"github.com/code-execution-engine/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Handler processes incoming tasks from the queue.
type Handler struct {
	redisClient      *redis.Client
	judgeRepo        *judge.Repository
	executionService *execution.Service
}

// NewHandler creates a new instance of Handler.
func NewHandler(redisClient *redis.Client, judgeRepo *judge.Repository, executionService *execution.Service) *Handler {
	return &Handler{
		redisClient:      redisClient,
		judgeRepo:        judgeRepo,
		executionService: executionService,
	}
}

// HandleRunCodeTask processes a run code task.
func (h *Handler) HandleRunCodeTask(ctx context.Context, t *asynq.Task) error {
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
	_ = h.judgeRepo.UpdateJobResult(jobUUID, judge.UpdateJobDTO{
		Status: types.StatusProcessing,
	})
	_ = queue.SetStatus(ctx, h.redisClient, payload.JobID, types.StatusProcessing)

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

	res := h.executionService.Execute(j)

	// Cache final result in Redis (polling optimization)
	err = queue.SetResult(ctx, h.redisClient, payload.JobID, res)
	if err != nil {
		telemetry.Error("Failed to save result to Redis", "job_id", payload.JobID, "error", err)
	} else {
		telemetry.Info("Saved execution result to Redis", "job_id", payload.JobID)
	}

	// Save final result in Postgres
	dbErr := h.judgeRepo.UpdateJobResult(jobUUID, judge.UpdateJobDTO{
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
}
