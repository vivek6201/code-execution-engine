package judge

import (
	"net/http"

	"github.com/code-execution-engine/internal/server/utility"
	"github.com/gin-gonic/gin"
)

// LanguageInfo represents a supported programming language.
type LanguageInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// StatusInfo represents an execution status.
type StatusInfo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// SupportedLanguages lists the sandbox compilers and runtimes.
var SupportedLanguages = []LanguageInfo{
	{ID: "python", Name: "Python", Version: "3.x"},
	{ID: "javascript", Name: "JavaScript (Node.js)", Version: "18.x"},
	{ID: "cpp", Name: "C++ (GCC)", Version: "G++ 12"},
	{ID: "java", Name: "Java (OpenJDK)", Version: "JDK 17"},
}

// ExecutionStatuses defines the set of response statuses.
var ExecutionStatuses = []StatusInfo{
	{ID: "QUEUED", Description: "Job is in queue waiting to be processed"},
	{ID: "PROCESSING", Description: "Job is being processed in the sandbox"},
	{ID: "SUCCESS", Description: "Job completed successfully and all test cases passed"},
	{ID: "FAILED", Description: "Job completed but some or all test cases failed"},
	{ID: "ERROR", Description: "Compile error, runtime error, or sandbox exception"},
	{ID: "TLE", Description: "Time Limit Exceeded"},
}

// GetLanguages handles GET /api/v1/languages
func GetLanguages(c *gin.Context) {
	utility.SuccessResponse(c, http.StatusOK, "Languages retrieved successfully", SupportedLanguages)
}

// GetStatuses handles GET /api/v1/statuses
func GetStatuses(c *gin.Context) {
	utility.SuccessResponse(c, http.StatusOK, "Statuses retrieved successfully", ExecutionStatuses)
}
