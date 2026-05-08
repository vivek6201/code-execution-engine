package judge

import (
	"errors"

	"github.com/code-execution-engine/internal/types"
)

// TestCaseDTO represents a test case in the API request.
type TestCaseDTO struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
}

// CreateJobRequest is the payload for submitting a code execution job.
type CreateJobRequest struct {
	Code      string        `json:"code" binding:"required"`
	Language  string        `json:"language" binding:"required"`
	Input     string        `json:"input"`
	TestCases []TestCaseDTO `json:"test_cases,omitempty" binding:"omitempty,dive"`
}

// Validate checks that the language is supported.
func (r *CreateJobRequest) Validate() error {
	switch r.Language {
	case "cpp", "python", "java", "javascript":
		return nil
	default:
		return errors.New("invalid language")
	}
}

// UpdateJobDTO is a strict type for updating a job's status and results.
type UpdateJobDTO struct {
	Status     types.Status           `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	FatalError string                 `json:"fatal_error,omitempty"`
	TimeMs     int64                  `json:"time_ms,omitempty"`
	MemoryKB   int64                  `json:"memory_kb,omitempty"`
	Passed     int                    `json:"passed,omitempty"`
	TestCases  []types.TestCaseResult `json:"test_cases,omitempty" gorm:"serializer:json"`
}
