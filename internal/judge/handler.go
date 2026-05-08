package judge

import (
	"net/http"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/code-execution-engine/internal/types"
	"github.com/code-execution-engine/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Handler handles HTTP requests for code execution jobs.
type Handler struct {
	queue       *queue.RedisQueue
	redisClient *redis.Client
}

// NewHandler creates a new judge HTTP handler.
func NewHandler(q *queue.RedisQueue, redisClient *redis.Client) *Handler {
	return &Handler{
		queue:       q,
		redisClient: redisClient,
	}
}

// RunCode handles POST /api/v1/run
func (h *Handler) RunCode(c *gin.Context) {
	var req CreateJobRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", utility.ParseError(err))
		return
	}

	if err := req.Validate(); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	j := types.Job{
		Code:     req.Code,
		Language: req.Language,
		Input:    req.Input,
	}
	for _, tc := range req.TestCases {
		j.TestCases = append(j.TestCases, types.TestCase{
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

// GetResult handles GET /api/v1/run/:id
func (h *Handler) GetResult(c *gin.Context) {
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
