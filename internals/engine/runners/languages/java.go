package languages

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/infra/isolation"
)

const (
	javaImage      = "openjdk:17.0.1-jdk-slim"
	javaFilename   = "Main.java"
	javaCompileCmd = "javac Main.java"
	javaRunCmd     = "java Main"
)

type JavaRunner struct {
	client *isolation.Client
}

func NewJavaRunner(client *isolation.Client) *JavaRunner {
	return &JavaRunner{client: client}
}

func (r *JavaRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, javaImage, javaFilename, req.Code, javaCompileCmd+" && "+javaRunCmd, req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{Output: res.Output, Error: res.Error}, nil
}

func (r *JavaRunner) RunBatch(ctx context.Context, req *runners.BatchRequest) ([]runners.ExecuteResult, error) {
	results, err := r.client.RunBatch(ctx, javaImage, javaFilename, req.Code, javaCompileCmd, javaRunCmd, req.Inputs)
	if err != nil {
		return nil, err
	}
	out := make([]runners.ExecuteResult, len(results))
	for i, res := range results {
		out[i] = runners.ExecuteResult{Output: res.Output, Error: res.Error}
	}
	return out, nil
}
