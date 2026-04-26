package languages

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/infra/isolation"
)

type PythonRunner struct {
	client *isolation.Client
}

func NewPythonRunner(client *isolation.Client) *PythonRunner {
	return &PythonRunner{
		client: client,
	}
}

func (r *PythonRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, "python:3.9-slim", "main.py", req.Code, "python3 main.py", req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{
		Output: res.Output,
		Error:  res.Error,
	}, nil
}
