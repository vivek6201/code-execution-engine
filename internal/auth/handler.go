package auth

import (
	"net/http"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the auth module.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", utility.ParseError(err))
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		utility.ErrorResponse(c, http.StatusConflict, "Registration failed", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", utility.ParseError(err))
		return
	}

	resp, err := h.service.Login(req)
	if err != nil {
		utility.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "Login successful", resp)
}

// GetMe handles GET /api/v1/dashboard/me
func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := h.service.GetUser(userID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "User retrieved", user)
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
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to create API key", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusCreated, "API key created. Save this key — it won't be shown again.", resp)
}

// ListKeys handles GET /api/v1/dashboard/keys
func (h *Handler) ListKeys(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	keys, err := h.service.ListAPIKeys(userID)
	if err != nil {
		utility.ErrorResponse(c, http.StatusInternalServerError, "Failed to list API keys", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "API keys retrieved", keys)
}

// RevokeKey handles DELETE /api/v1/dashboard/keys/:id
func (h *Handler) RevokeKey(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utility.ErrorResponse(c, http.StatusBadRequest, "Invalid key ID", err.Error())
		return
	}

	if err := h.service.RevokeAPIKey(userID, keyID); err != nil {
		utility.ErrorResponse(c, http.StatusNotFound, "API key not found", err.Error())
		return
	}

	utility.SuccessResponse(c, http.StatusOK, "API key revoked", nil)
}
