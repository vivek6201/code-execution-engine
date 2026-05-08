package languages

import (
	"context"

	"github.com/code-execution-engine/internal/engine/runners"
	"github.com/code-execution-engine/internal/isolation"
)

const (
	cppImage      = "gcc:latest"
	cppFilename   = "main.cpp"
	cppCompileCmd = "g++ main.cpp -o main"
	cppRunCmd     = "./main"
)

type CppRunner struct {
	client *isolation.Client
}

func NewCppRunner(client *isolation.Client) *CppRunner {
	return &CppRunner{client: client}
}

func (r *CppRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, cppImage, cppFilename, req.Code, cppCompileCmd+" && "+cppRunCmd, req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error}, nil
}

func (r *CppRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, error) {
	results, err := r.client.RunBatch(ctx, cppImage, cppFilename, req.Code, cppCompileCmd, cppRunCmd, req.Inputs)
	if err != nil {
		return nil, err
	}
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error}
	}
	return out, nil
}
