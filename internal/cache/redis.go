package cache

import "github.com/redis/go-redis/v9"

// NewRedisClient creates and returns a shared Redis client from the given URL.
func NewRedisClient(redisURL string) *redis.Client {
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{Addr: "localhost:6379"}
	}
	return redis.NewClient(opts)
}
