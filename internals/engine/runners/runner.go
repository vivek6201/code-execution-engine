package runners

import "context"

type ExecuteRequest struct {
	Code  string
	Input string
}

type ExecuteResult struct {
	Output string
	Error  string
}

type Runner interface {
	Run(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error)
}
