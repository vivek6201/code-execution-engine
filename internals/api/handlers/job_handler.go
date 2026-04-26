package handlers

import (
	"net/http"

	"github.com/code-execution-engine/internals/api/dtos"
	"github.com/code-execution-engine/internals/core/job"
	"github.com/code-execution-engine/internals/infra/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobHandler struct {
	queue *queue.RedisQueue
}

func NewJobHandler(q *queue.RedisQueue) *JobHandler {
	return &JobHandler{
		queue: q,
	}
}

func (h *JobHandler) RunCode(c *gin.Context) {
	var req dtos.CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	j := job.Job{
		Code:     req.Code,
		Language: req.Language,
		Input:    req.Input,
	}
	for _, tc := range req.TestCases {
		j.TestCases = append(j.TestCases, job.TestCase{
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	jobID := uuid.New().String()
	if err := h.queue.Enqueue(c.Request.Context(), jobID, j); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": "Job queued successfully",
	})
}

func (h *JobHandler) GetResult(c *gin.Context) {
	jobID := c.Param("id")
	res, err := h.queue.GetResult(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "PENDING"})
		return
	}

	c.JSON(http.StatusOK, res)
}
