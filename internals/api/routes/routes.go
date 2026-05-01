package routes

import (
	"github.com/code-execution-engine/internals/api/handlers"
	"github.com/code-execution-engine/internals/infra/queue"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(r *gin.Engine, q *queue.RedisQueue, redisClient *redis.Client) {
	h := handlers.NewJobHandler(q, redisClient)

	api := r.Group("/api/v1")
	{
		api.POST("/run", h.RunCode)
		api.GET("/run/:id", h.GetResult)
	}
}
