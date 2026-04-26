package handlers

import (
	"net/http"

	"github.com/code-execution-engine/internals/api/dtos"
	"github.com/code-execution-engine/internals/core/job"
	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	service *execution.Service
}

func NewJobHandler(s *execution.Service) *JobHandler {
	return &JobHandler{
		service: s,
	}
}

func (h *JobHandler) RunCode(c *gin.Context) {
	var req dtos.CreateJobRequest

	// bind + validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// convert dtos request into domain request
	j := job.Job{
		Code:     req.Code,
		Language: req.Language,
		Input:    req.Input,
	}

	result := h.service.Execute(j)

	c.JSON(http.StatusOK, result)
}
