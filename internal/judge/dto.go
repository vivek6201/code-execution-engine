package judge

import "errors"

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
