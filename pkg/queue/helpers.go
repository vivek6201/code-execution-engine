package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/code-execution-engine/internal/types"
	"github.com/redis/go-redis/v9"
)

// Redis key prefixes and constants
const (
	JobQueueKey     = "job_queue"
	JobResultPrefix = "job_result:"
	ResultTTL       = time.Hour
)

// SetStatus stores a status-only result for a job (used for QUEUED → PROCESSING transitions).
func SetStatus(ctx context.Context, client *redis.Client, id string, status types.Status) error {
	r := types.Result{Status: status}
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return client.Set(ctx, JobResultPrefix+id, bytes, ResultTTL).Err()
}

// SetResult stores the final execution result for a job.
func SetResult(ctx context.Context, client *redis.Client, id string, r types.Result) error {
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return client.Set(ctx, JobResultPrefix+id, bytes, ResultTTL).Err()
}

// GetResult retrieves the current status/result for a job.
func GetResult(ctx context.Context, client *redis.Client, id string) (*types.Result, error) {
	val, err := client.Get(ctx, JobResultPrefix+id).Result()
	if err != nil {
		return nil, err
	}
	var r types.Result
	if err := json.Unmarshal([]byte(val), &r); err != nil {
		return nil, err
	}
	return &r, nil
}
