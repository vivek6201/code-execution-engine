package result

type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
	StatusError   Status = "ERROR"
	StatusTLE     Status = "TLE"
)

type TestCaseResult struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	Status         Status `json:"status"`
	Error          string `json:"error,omitempty"`
}

type Result struct {
	Status     Status           `json:"status"`
	Output     string           `json:"output,omitempty"`
	Error      string           `json:"error,omitempty"`
	TestCases  []TestCaseResult `json:"test_cases,omitempty"`
	Total      int              `json:"total,omitempty"`
	Passed     int              `json:"passed,omitempty"`
	FatalError string           `json:"fatal_error,omitempty"`
}
