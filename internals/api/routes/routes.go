package routes

import (
	"github.com/code-execution-engine/internals/api/handlers"
	"github.com/code-execution-engine/internals/core/execution"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, s *execution.Service) {
	h := handlers.NewJobHandler(s)

	api := r.Group("/api/v1")
	{
		api.POST("/run", h.RunCode)
	}
}
