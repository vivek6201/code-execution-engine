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
