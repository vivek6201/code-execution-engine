package handlers

import (
	"net/http"

	"github.com/code-execution-engine/internals/api/dtos"
	"github.com/code-execution-engine/internals/api/utility"
	"github.com/code-execution-engine/internals/core/job"
	"github.com/code-execution-engine/internals/infra/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JobHandler struct {
	queue       *queue.RedisQueue
	redisClient *redis.Client
}

func NewJobHandler(q *queue.RedisQueue, redisClient *redis.Client) *JobHandler {
	return &JobHandler{
		queue:       q,
		redisClient: redisClient,
	}
}

func (h *JobHandler) RunCode(c *gin.Context) {
	var req dtos.CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
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
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue job", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusAccepted, "Job queued successfully", gin.H{
		"job_id": jobID,
	})
}

func (h *JobHandler) GetResult(c *gin.Context) {
	jobID := c.Param("id")
	res, err := queue.GetResult(c.Request.Context(), h.redisClient, jobID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusNotFound, "Job not found", "no job found with the given ID")
		return
	}

	// Determine message based on status
	msg := "Result fetched successfully"
	switch res.Status {
	case "QUEUED":
		msg = "Job is queued for processing"
	case "PROCESSING":
		msg = "Job is being processed"
	}

	utility.SuccessResponse(c, http.StatusOK, msg, res)
}
