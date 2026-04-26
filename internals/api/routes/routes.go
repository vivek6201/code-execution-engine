package routes

import (
	"github.com/code-execution-engine/internals/api/handlers"
	"github.com/code-execution-engine/internals/infra/queue"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, q *queue.RedisQueue) {
	h := handlers.NewJobHandler(q)

	api := r.Group("/api/v1")
	{
		api.POST("/run", h.RunCode)
		api.GET("/run/:id", h.GetResult)
	}
}
