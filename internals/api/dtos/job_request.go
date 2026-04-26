package dtos

import "errors"

type CreateJobRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
	Input    string `json:"input" binding:"required"`
}

func (r *CreateJobRequest) Validate() error {
	switch r.Language {
	case "cpp", "python", "java":
		return nil
	default:
		return errors.New("invalid language")
	}
}
