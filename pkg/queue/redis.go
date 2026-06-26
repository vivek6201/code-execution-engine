package queue

import (
	"context"
	"encoding/json"

	"github.com/code-execution-engine/internal/types"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const (
	TaskRunCode = "task:run_code"
)

// RunCodePayload defines the payload structure sent to the worker via Asynq.
type RunCodePayload struct {
	JobID         string           `json:"job_id"`
	Code          string           `json:"code"`
	Language      string           `json:"language"`
	Input         string           `json:"input"`
	TestCases     []types.TestCase `json:"test_cases"`
	TimeLimitMS   *int64           `json:"time_limit_ms,omitempty"`
	MemoryLimitKB *int64           `json:"memory_limit_kb,omitempty"`
	CallbackURL   *string          `json:"callback_url,omitempty"`
}

type RedisQueue struct {
	client      *redis.Client
	asynqClient *asynq.Client
}

// NewRedisQueue creates a queue service using an existing Redis client and Redis URL.
func NewRedisQueue(client *redis.Client, redisURL string) *RedisQueue {
	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		redisOpt = asynq.RedisClientOpt{Addr: "localhost:6379"}
	}
	return &RedisQueue{
		client:      client,
		asynqClient: asynq.NewClient(redisOpt),
	}
}

// Client exposes the underlying Redis client for reuse.
func (q *RedisQueue) Client() *redis.Client {
	return q.client
}

// Close closes the Asynq client connections.
func (q *RedisQueue) Close() error {
	return q.asynqClient.Close()
}

// Enqueue serializes the code job and pushes it as an Asynq task.
func (q *RedisQueue) Enqueue(ctx context.Context, id string, j types.Job, timeLimit *int64, memLimit *int64, callbackURL *string) error {
	payload := RunCodePayload{
		JobID:         id,
		Code:          j.Code,
		Language:      j.Language,
		Input:         j.Input,
		TestCases:     j.TestCases,
		TimeLimitMS:   timeLimit,
		MemoryLimitKB: memLimit,
		CallbackURL:   callbackURL,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Store initial status as QUEUED in Redis cache for fast polling
	if err := SetStatus(ctx, q.client, id, types.StatusQueued); err != nil {
		return err
	}

	task := asynq.NewTask(TaskRunCode, bytes)
	_, err = q.asynqClient.EnqueueContext(ctx, task)
	return err
}
