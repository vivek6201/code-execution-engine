package languages

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/infra/isolation"
)

const (
	pythonImage    = "python:3.9-slim"
	pythonFilename = "main.py"
	pythonRunCmd   = "python3 main.py"
)

type PythonRunner struct {
	client *isolation.Client
}

func NewPythonRunner(client *isolation.Client) *PythonRunner {
	return &PythonRunner{client: client}
}

func (r *PythonRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, pythonImage, pythonFilename, req.Code, pythonRunCmd, req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error}, nil
}

func (r *PythonRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, error) {
	results, err := r.client.RunBatch(ctx, pythonImage, pythonFilename, req.Code, "", pythonRunCmd, req.Inputs)
	if err != nil {
		return nil, err
	}
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error}
	}
	return out, nil
}
