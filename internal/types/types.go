// Package types defines shared domain types used across modules.
// This package has zero internal dependencies to prevent import cycles.
package types

// TestCase holds a single test case with input and expected output.
type TestCase struct {
	Input          string
	ExpectedOutput string
}

// Job represents a code execution request.
type Job struct {
	Code          string
	Language      string
	Input         string
	TestCases     []TestCase
	TimeLimitMS   int64
	MemoryLimitKB int64
}

// Status represents the lifecycle state of a job or test case.
type Status string

const (
	StatusQueued     Status = "QUEUED"
	StatusProcessing Status = "PROCESSING"
	StatusSuccess    Status = "SUCCESS"
	StatusFailed     Status = "FAILED"
	StatusError      Status = "ERROR"
	StatusTLE        Status = "TLE"
)

// TestCaseResult holds the outcome of a single test case execution.
type TestCaseResult struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	Status         Status `json:"status"`
	Error          string `json:"error,omitempty"`
	TimeMs         int64  `json:"time_ms"`
}

// Result holds the overall execution outcome for a job.
type Result struct {
	Status     Status           `json:"status"`
	Output     string           `json:"output,omitempty"`
	Error      string           `json:"error,omitempty"`
	TestCases  []TestCaseResult `json:"test_cases,omitempty"`
	Total      int              `json:"total,omitempty"`
	Passed     int              `json:"passed,omitempty"`
	FatalError string           `json:"fatal_error,omitempty"`
	TimeMs     int64            `json:"time_ms,omitempty"`
	MemoryKB   int64            `json:"memory_kb,omitempty"`
}
