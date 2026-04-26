package languages

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/infra/isolation"
)

type JavaRunner struct {
	client *isolation.Client
}

func NewJavaRunner(client *isolation.Client) *JavaRunner {
	return &JavaRunner{
		client: client,
	}
}

func (r *JavaRunner) Run(ctx context.Context, req *runners.ExecuteRequest) (*runners.ExecuteResult, error) {
	res, err := r.client.Run(ctx, "openjdk:17.0.1-jdk-slim", "Main.java", req.Code, "javac Main.java && java Main", req.Input)
	if err != nil {
		return nil, err
	}
	return &runners.ExecuteResult{
		Output: res.Output,
		Error:  res.Error,
	}, nil
}
