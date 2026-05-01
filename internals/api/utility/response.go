package utility

import "github.com/gin-gonic/gin"

// APIResponse is the standard success response envelope.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse is the standard error response envelope.
type ErrorRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// Success sends a standardised success JSON response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends a standardised error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, ErrorRes{
		Success: false,
		Message: message,
		Error:   err,
	})
}
