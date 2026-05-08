package languages

import (
	"context"

	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/isolation"
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

func (r *CppRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, int64, error) {
	res, memKB, err := r.client.Run(ctx, cppImage, cppFilename, req.Code, cppCompileCmd+" && "+cppRunCmd, req.Input)
	if err != nil {
		return nil, 0, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}, memKB, nil
}

func (r *CppRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, int64, error) {
	results, memKB, err := r.client.RunBatch(ctx, cppImage, cppFilename, req.Code, cppCompileCmd, cppRunCmd, req.Inputs)
	if err != nil {
		return nil, 0, err
	}
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}
	}
	return out, memKB, nil
}
