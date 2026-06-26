package executor

import (
	"context"

	"github.com/code-execution-engine/internal/judge/engine/runners"
)

type Executor struct {
	factory *runners.Factory
}

func NewExecutor(factory *runners.Factory) *Executor {
	return &Executor{
		factory: factory,
	}
}

func (e *Executor) Run(language, code, input string, timeLimitMS int64, memoryLimitKB int64) (runners.ExecuteResult, int64) {
	r, err := e.factory.GetRunner(language)

	if err != nil {
		return runners.ExecuteResult{
			Error: "unsupported language",
		}, 0
	}

	ctx := context.Background()

	res, memKB, err := r.Run(ctx, &runners.ExecuteRequest{
		Code:          code,
		Input:         input,
		TimeLimitMS:   timeLimitMS,
		MemoryLimitKB: memoryLimitKB,
	})

	if err != nil {
		if err.Error() == "timeout" {
			return runners.ExecuteResult{Error: "timeout"}, memKB
		}
		return runners.ExecuteResult{Error: err.Error()}, memKB
	}
	return *res, memKB
}

// RunBatch executes code once and runs it against multiple inputs in a single container.
// Accepts a context for cancellation support (early exit on failure).
func (e *Executor) RunBatch(ctx context.Context, language, code string, inputs []string, timeLimitMS int64, memoryLimitKB int64) ([]runners.ExecuteResult, int64, error) {
	r, err := e.factory.GetRunner(language)
	if err != nil {
		return nil, 0, err
	}

	return r.RunBatch(ctx, &runners.BatchRequest{
		Code:          code,
		Inputs:        inputs,
		TimeLimitMS:   timeLimitMS,
		MemoryLimitKB: memoryLimitKB,
	})
}
