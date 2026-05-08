package languages

import (
	"context"

	"github.com/code-execution-engine/internal/engine/runners"
	"github.com/code-execution-engine/internal/isolation"
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

func (r *NodeRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, image, filename, req.Code, runCmd, req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error}, nil
}

func (r *NodeRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, error) {
	results, err := r.client.RunBatch(ctx, image, filename, req.Code, compileCmd, runCmd, req.Inputs)
	if err != nil {
		return nil, err
	}
	// Convert isolation results to runners results
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error}
	}
	return out, nil
}
