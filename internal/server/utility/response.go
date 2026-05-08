package utility

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// APIResponse is the standard success response envelope.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorRes is the standard error response envelope.
type ErrorRes struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   any    `json:"error"`
}

// SuccessResponse sends a standardised success JSON response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends a standardised error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, message string, err any) {
	c.JSON(statusCode, ErrorRes{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// ParseError formats the error to be more readable. If it's a validation error,
// it returns a map of fields and their validation failure reasons.
func ParseError(err error) any {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range validationErrs {
			errors[e.Field()] = getErrorMessage(e)
		}
		return errors
	}
	if err != nil {
		return err.Error()
	}
	return "unknown error"
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address format"
	case "min":
		return "Must be at least " + e.Param() + " characters"
	case "max":
		return "Must be at most " + e.Param() + " characters"
	default:
		return "Validation failed on '" + e.Tag() + "'"
	}
}
