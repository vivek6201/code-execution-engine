package auth

import (
	"net/http"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for user authentication.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler.
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
