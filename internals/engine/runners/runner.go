package runners

import "context"

type ExecuteRequest struct {
	Code  string
	Input string
}

type BatchRequest struct {
	Code   string
	Inputs []string
}

type ExecuteResult struct {
	Output string
	Error  string
}

type Runner interface {
	Run(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error)
	RunBatch(ctx context.Context, req *BatchRequest) ([]ExecuteResult, error)
}
