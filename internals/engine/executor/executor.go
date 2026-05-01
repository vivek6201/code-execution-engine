package executor

import (
	"context"

	"github.com/code-execution-engine/internals/engine/runners"
)

type Executor struct {
	factory *runners.Factory
}

func NewExecutor(factory *runners.Factory) *Executor {
	return &Executor{
		factory: factory,
	}
}

func (e *Executor) Run(language, code, input string) runners.ExecuteResult {
	r, err := e.factory.GetRunner(language)

	if err != nil {
		return runners.ExecuteResult{
			Error: "unsupported language",
		}
	}

	ctx := context.Background()

	res, err := r.Run(ctx, &runners.ExecuteRequest{
		Code:  code,
		Input: input,
	})

	if err != nil {
		if err.Error() == "timeout" {
			return runners.ExecuteResult{Error: "timeout"}
		}
		return runners.ExecuteResult{Error: err.Error()}
	}
	return *res
}

// RunBatch executes code once and runs it against multiple inputs in a single container.
// Accepts a context for cancellation support (early exit on failure).
func (e *Executor) RunBatch(ctx context.Context, language, code string, inputs []string) ([]runners.ExecuteResult, error) {
	r, err := e.factory.GetRunner(language)
	if err != nil {
		return nil, err
	}

	return r.RunBatch(ctx, &runners.BatchRequest{
		Code:   code,
		Inputs: inputs,
	})
}
