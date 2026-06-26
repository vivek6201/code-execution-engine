package runners

import "context"

type ExecuteRequest struct {
	Code          string
	Input         string
	TimeLimitMS   int64
	MemoryLimitKB int64
}

type BatchRequest struct {
	Code          string
	Inputs        []string
	TimeLimitMS   int64
	MemoryLimitKB int64
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
