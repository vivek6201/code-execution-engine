package languages

import (
	"context"

	"github.com/code-execution-engine/internal/judge/engine/runners"
	"github.com/code-execution-engine/internal/judge/isolation"
)

const (
	image      = "node:22.18-alpine"
	filename   = "main.js"
	compileCmd = ""
	runCmd     = "node main.js"
)

type NodeRunner struct {
	client *isolation.Client
}

func NewNodeRunner(client *isolation.Client) *NodeRunner {
	return &NodeRunner{client: client}
}

func (r *NodeRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, int64, error) {
	res, memKB, err := r.client.Run(ctx, "", image, filename, req.Code, compileCmd, runCmd, nil, req.Input, req.MemoryLimitKB*1024, req.TimeLimitMS)
	if err != nil {
		return nil, 0, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}, memKB, nil
}

func (r *NodeRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, int64, error) {
	results, memKB, err := r.client.RunBatch(ctx, "", image, filename, req.Code, compileCmd, runCmd, nil, req.Inputs, req.MemoryLimitKB*1024, req.TimeLimitMS)
	if err != nil {
		return nil, 0, err
	}
	// Convert isolation results to runners results
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error, TimeMs: res.TimeMs}
	}
	return out, memKB, nil
}
