package languages

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/infra/isolation"
)

type CppRunner struct {
	client *isolation.Client
}

func NewCppRunner(client *isolation.Client) *CppRunner {
	return &CppRunner{
		client: client,
	}
}

func (r *CppRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, "gcc:latest", "main.cpp", req.Code, "g++ main.cpp -o main && ./main", req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{
		Output: res.Output,
		Error:  res.Error,
	}, nil
}
