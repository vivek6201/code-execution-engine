package execution

import (
	"context"
	"sync"

	"github.com/code-execution-engine/internal/engine/evaluator"
	"github.com/code-execution-engine/internal/types"
	"github.com/code-execution-engine/internal/engine/executor"
)

type Service struct {
	exec *executor.Executor
	eval *evaluator.Service
}

func NewService(e *executor.Executor, eval *evaluator.Service) *Service {
	return &Service{exec: e, eval: eval}
}

func (s *Service) Execute(j types.Job) types.Result {
	if j.Code == "" || j.Language == "" {
		return types.Result{Status: types.StatusError, FatalError: "invalid job"}
	}

	// Single execution flow (no test cases)
	if len(j.TestCases) == 0 {
		res := s.exec.Run(j.Language, j.Code, j.Input)
		if res.Error != "" {
			status := types.StatusError
			if res.Error == "timeout" {
				status = types.StatusTLE
			}
			return types.Result{
				Status: status,
				Output: res.Output,
				Error:  res.Error,
			}
		}
		return types.Result{
			Status: types.StatusSuccess,
			Output: res.Output,
		}
	}

	// Batch test case execution — compile once, run N times concurrently
	return s.executeBatch(j)
}

func (s *Service) executeBatch(j types.Job) types.Result {
	inputs := make([]string, len(j.TestCases))
	for i, tc := range j.TestCases {
		inputs[i] = tc.Input
	}

	// Create a cancellable context — we cancel on first failure to stop
	// remaining test case executions early.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run all test cases concurrently in a single container
	batchResults, err := s.exec.RunBatch(ctx, j.Language, j.Code, inputs)

	finalResult := types.Result{
		TestCases: make([]types.TestCaseResult, 0, len(j.TestCases)),
		Total:     len(j.TestCases),
	}

	if err != nil {
		return types.Result{
			Status:     types.StatusError,
			FatalError: err.Error(),
		}
	}

	// Evaluate results concurrently and cancel on first failure.
	// Since batchResults are indexed, we evaluate in parallel but build
	// the final result slice in order.
	tcResults := make([]types.TestCaseResult, len(j.TestCases))
	var passed int
	var mu sync.Mutex
	var failed bool
	var wg sync.WaitGroup

	for i, tc := range j.TestCases {
		wg.Add(1)
		go func(idx int, tc types.TestCase) {
			defer wg.Done()

			res := batchResults[idx]
			tcRes := types.TestCaseResult{
				Input:          tc.Input,
				ExpectedOutput: tc.ExpectedOutput,
			}

			if res.Error == "cancelled" {
				tcRes.Status = types.StatusError
				tcRes.Error = "skipped: previous test case failed"
			} else if res.Error != "" {
				if res.Error == "timeout" {
					tcRes.Status = types.StatusTLE
				} else {
					tcRes.Status = types.StatusError
					tcRes.Error = res.Error
				}
				// Cancel remaining on error/TLE
				mu.Lock()
				failed = true
				mu.Unlock()
				cancel()
			} else {
				tcRes.ActualOutput = res.Output
				if s.eval.Evaluate(res.Output, tc.ExpectedOutput) {
					tcRes.Status = types.StatusSuccess
					mu.Lock()
					passed++
					mu.Unlock()
				} else {
					tcRes.Status = types.StatusFailed
					// Cancel remaining on wrong answer
					mu.Lock()
					failed = true
					mu.Unlock()
					cancel()
				}
			}

			tcResults[idx] = tcRes
		}(i, tc)
	}

	wg.Wait()

	finalResult.TestCases = tcResults
	finalResult.Passed = passed

	if !failed && passed == finalResult.Total {
		finalResult.Status = types.StatusSuccess
	} else {
		finalResult.Status = types.StatusFailed
	}

	return finalResult
}
