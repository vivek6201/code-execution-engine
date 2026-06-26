package languages

import (
	"context"

	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/isolation"
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

func (r *PythonRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, int64, error) {
	res, memKB, err := r.client.Run(ctx, "", pythonImage, pythonFilename, req.Code, "", pythonRunCmd, nil, req.Input, req.MemoryLimitKB*1024, req.TimeLimitMS)
	if err != nil {
		return nil, 0, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}, memKB, nil
}

func (r *PythonRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, int64, error) {
	results, memKB, err := r.client.RunBatch(ctx, "", pythonImage, pythonFilename, req.Code, "", pythonRunCmd, nil, req.Inputs, req.MemoryLimitKB*1024, req.TimeLimitMS)
	if err != nil {
		return nil, 0, err
	}
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}
	}
	return out, memKB, nil
}
