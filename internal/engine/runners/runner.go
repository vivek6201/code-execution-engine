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
	TimeMs int64
}

type Runner interface {
	Run(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, int64, error)
	RunBatch(ctx context.Context, req *BatchRequest) ([]ExecuteResult, int64, error)
}
