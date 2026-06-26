package dashboard

import (
	"net/http"
	"strconv"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for dashboard operations.
type Handler struct {
	service *Service
}

// NewHandler creates a new dashboard handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe handles GET /api/v1/dashboard/me
func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.service.GetUser(userID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", user)
}

// UpdatePlan handles PATCH /api/v1/dashboard/plan
func (h *Handler) UpdatePlan(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", utility.ParseError(err))
		return
	}

	if err := h.service.UpdatePlan(userID, req.Plan); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Failed to update plan", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "Plan updated successfully", nil)
}

// CreateKey handles POST /api/v1/dashboard/keys
func (h *Handler) CreateKey(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", utility.ParseError(err))
		return
	}

	resp, err := h.service.GenerateAPIKey(userID, req.Name)
	if err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate API key", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusCreated, "API key created successfully", resp)
}

// ListKeys handles GET /api/v1/dashboard/keys
func (h *Handler) ListKeys(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	keys, err := h.service.ListAPIKeys(userID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch API keys", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "API keys retrieved successfully", keys)
}

// RevokeKey handles DELETE /api/v1/dashboard/keys/:id
func (h *Handler) RevokeKey(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid key ID", nil)
		return
	}

	if err := h.service.RevokeAPIKey(userID, keyID); err != nil {
		utility.ErrorResponse(c, http.StatusNotFound, "API key not found or already revoked", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "API key revoked successfully", nil)
}

// GetJobs handles GET /api/v1/dashboard/jobs
func (h *Handler) GetJobs(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	lang := c.Query("language")
	status := c.Query("status")
	apiKeyID := c.Query("api_key_id")

	resp, err := h.service.GetJobs(userID, page, limit, lang, status, apiKeyID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve job log history", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "Job logs retrieved successfully", resp)
}

// GetMetrics handles GET /api/v1/dashboard/metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	resp, err := h.service.GetMetrics(userID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve dashboard metrics", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "Dashboard metrics retrieved successfully", resp)
}
