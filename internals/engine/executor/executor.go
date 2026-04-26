package executor

import "github.com/code-execution-engine/internals/engine/runners"

type Executor struct {
	factory *runners.Factory
}

func NewExecutor(factory *runners.Factory) *Executor {
	return &Executor{
		factory: factory,
	}
}
