package queue

import (
	"context"
	"encoding/json"

	"github.com/code-execution-engine/internal/types"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a queue service using an existing Redis client.
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// Client exposes the underlying Redis client for reuse (e.g., rate limiter).
func (q *RedisQueue) Client() *redis.Client {
	return q.client
}

type queuedJob struct {
	ID  string    `json:"id"`
	Job types.Job `json:"job"`
}

// Enqueue pushes a job onto the queue and sets its initial status to QUEUED.
func (q *RedisQueue) Enqueue(ctx context.Context, id string, j types.Job) error {
	qj := queuedJob{ID: id, Job: j}
	bytes, err := json.Marshal(qj)
	if err != nil {
		return err
	}

	// Set initial status to QUEUED before pushing to the queue
	if err := SetStatus(ctx, q.client, id, types.StatusQueued); err != nil {
		return err
	}

	return q.client.LPush(ctx, JobQueueKey, bytes).Err()
}

// Dequeue blocks until a job is available, pops it, and updates status to PROCESSING.
func (q *RedisQueue) Dequeue(ctx context.Context) (string, types.Job, error) {
	res, err := q.client.BRPop(ctx, 0, JobQueueKey).Result()
	if err != nil {
		return "", types.Job{}, err
	}
	var qj queuedJob
	if err := json.Unmarshal([]byte(res[1]), &qj); err != nil {
		return "", types.Job{}, err
	}

	// Update status to PROCESSING as soon as a worker picks up the job
	_ = SetStatus(ctx, q.client, qj.ID, types.StatusProcessing)

	return qj.ID, qj.Job, nil
}
