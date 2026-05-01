package dtos

import "errors"

type TestCase struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
}

type CreateJobRequest struct {
	Code      string     `json:"code" binding:"required"`
	Language  string     `json:"language" binding:"required"`
	Input     string     `json:"input"`
	TestCases []TestCase `json:"test_cases,omitempty" binding:"omitempty,dive"`
}

func (r *CreateJobRequest) Validate() error {
	switch r.Language {
	case "cpp", "python", "java":
		return nil
	default:
		return errors.New("invalid language")
	}
}
