package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/code-execution-engine/internals/core/job"
	"github.com/code-execution-engine/internals/core/result"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(redisURL string) *RedisQueue {
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{Addr: "localhost:6379"}
	}
	return &RedisQueue{
		client: redis.NewClient(opts),
	}
}

type queuedJob struct {
	ID  string  `json:"id"`
	Job job.Job `json:"job"`
}

func (q *RedisQueue) Enqueue(ctx context.Context, id string, j job.Job) error {
	qj := queuedJob{ID: id, Job: j}
	bytes, err := json.Marshal(qj)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, "job_queue", bytes).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (string, job.Job, error) {
	res, err := q.client.BRPop(ctx, 0, "job_queue").Result()
	if err != nil {
		return "", job.Job{}, err
	}
	var qj queuedJob
	if err := json.Unmarshal([]byte(res[1]), &qj); err != nil {
		return "", job.Job{}, err
	}
	return qj.ID, qj.Job, nil
}

func (q *RedisQueue) SetResult(ctx context.Context, id string, r result.Result) error {
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return q.client.Set(ctx, "job_result:"+id, bytes, time.Hour).Err()
}

func (q *RedisQueue) GetResult(ctx context.Context, id string) (*result.Result, error) {
	val, err := q.client.Get(ctx, "job_result:"+id).Result()
	if err != nil {
		return nil, err
	}
	var r result.Result
	if err := json.Unmarshal([]byte(val), &r); err != nil {
		return nil, err
	}
	return &r, nil
}
