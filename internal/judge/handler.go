package judge

import (
	"net/http"
	"strings"

	"github.com/code-execution-engine/internal/models"
	"github.com/code-execution-engine/internal/server/utility"
	"github.com/code-execution-engine/internal/types"
	"github.com/code-execution-engine/pkg/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Handler handles HTTP requests for code execution jobs.
type Handler struct {
	queue       *queue.RedisQueue
	redisClient *redis.Client
	repo        *Repository
}

// NewHandler creates a new judge HTTP handler.
func NewHandler(q *queue.RedisQueue, redisClient *redis.Client, repo *Repository) *Handler {
	return &Handler{
		queue:       q,
		redisClient: redisClient,
		repo:        repo,
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

	userID := c.MustGet("user_id").(uuid.UUID)
	apiKey := c.MustGet("api_key").(*models.APIKey)
	planLimits := c.MustGet("plan_limits").(models.PlanLimits)

	// Validate and default custom limits
	var timeLimitVal int64 = planLimits.MaxTimeLimitMS
	if req.TimeLimitMS != nil {
		if *req.TimeLimitMS <= 0 || *req.TimeLimitMS > planLimits.MaxTimeLimitMS {
			utility.ErrorResponse(c, http.StatusBadRequest, "Validation failed", "time_limit_ms exceeds your plan limits")
			return
		}
		timeLimitVal = *req.TimeLimitMS
	}

	var memLimitVal int64 = planLimits.MaxMemoryLimitKB
	if req.MemoryLimitKB != nil {
		if *req.MemoryLimitKB <= 0 || *req.MemoryLimitKB > planLimits.MaxMemoryLimitKB {
			utility.ErrorResponse(c, http.StatusBadRequest, "Validation failed", "memory_limit_kb exceeds your plan limits")
			return
		}
		memLimitVal = *req.MemoryLimitKB
	}

	if req.CallbackURL != nil && *req.CallbackURL != "" {
		if !strings.HasPrefix(*req.CallbackURL, "http://") && !strings.HasPrefix(*req.CallbackURL, "https://") {
			utility.ErrorResponse(c, http.StatusBadRequest, "Validation failed", "invalid callback_url format")
			return
		}
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

	jobID := uuid.New()

	jobRecord := JobRecord{
		ID:            jobID,
		UserID:        userID,
		APIKeyID:      apiKey.ID,
		Language:      req.Language,
		Code:          req.Code,
		Status:        types.StatusQueued,
		Total:         len(req.TestCases),
		TimeLimitMS:   &timeLimitVal,
		MemoryLimitKB: &memLimitVal,
		CallbackURL:   req.CallbackURL,
	}
	if err := h.repo.CreateJob(&jobRecord); err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to create job record", err.Error())
		return
	}

	if err := h.queue.Enqueue(c.Request.Context(), jobID.String(), j, &timeLimitVal, &memLimitVal, req.CallbackURL); err != nil {
		// Try to mark job as failed in DB
		_ = h.repo.UpdateJobResult(jobID, UpdateJobDTO{
			Status:     types.StatusError,
			FatalError: "failed to enqueue",
		})
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue job", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusAccepted, "Job queued successfully", gin.H{
		"job_id": jobID.String(),
	})
}

// GetResult handles GET /api/v1/run/:id
func (h *Handler) GetResult(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobUUID, err := uuid.Parse(jobIDStr)
	if err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID format", err.Error())
		return
	}

	// 1. First try Redis for fast live polling
	res, err := queue.GetResult(c.Request.Context(), h.redisClient, jobIDStr)
	if err == nil {
		// Determine message based on status
		msg := "Result fetched successfully"
		switch res.Status {
		case "QUEUED":
			msg = "Job is queued for processing"
		case "PROCESSING":
			msg = "Job is being processed"
		}
		utility.SuccessResponse(c, http.StatusOK, msg, res)
		return
	}

	// 2. If Redis expired or doesn't have it, fetch from Postgres
	jobRecord, err := h.repo.GetJobByID(jobUUID)
	if err != nil || jobRecord == nil {
		utility.ErrorResponse(c, http.StatusNotFound, "Job not found", "no job found with the given ID")
		return
	}

	result := types.Result{
		Status:     jobRecord.Status,
		Output:     jobRecord.Output,
		Error:      jobRecord.Error,
		FatalError: jobRecord.FatalError,
		TimeMs:     jobRecord.TimeMs,
		MemoryKB:   jobRecord.MemoryKB,
		Total:      jobRecord.Total,
		Passed:     jobRecord.Passed,
		TestCases:  jobRecord.TestCases,
	}

	msg := "Result fetched successfully from database"
	switch result.Status {
	case types.StatusQueued:
		msg = "Job is queued for processing"
	case types.StatusProcessing:
		msg = "Job is being processed"
	}

	// Cache back to Redis for faster retrieval next time
	_ = queue.SetResult(c.Request.Context(), h.redisClient, jobIDStr, result)

	utility.SuccessResponse(c, http.StatusOK, msg, result)
}
