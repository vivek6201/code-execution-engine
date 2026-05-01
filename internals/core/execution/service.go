package execution

import (
	"context"
	"sync"

	"github.com/code-execution-engine/internals/core/evaluator"
	"github.com/code-execution-engine/internals/core/job"
	"github.com/code-execution-engine/internals/core/result"
	"github.com/code-execution-engine/internals/engine/executor"
)

type Service struct {
	exec *executor.Executor
	eval *evaluator.Service
}

func NewService(e *executor.Executor, eval *evaluator.Service) *Service {
	return &Service{exec: e, eval: eval}
}

func (s *Service) Execute(j job.Job) result.Result {
	if j.Code == "" || j.Language == "" {
		return result.Result{Status: result.StatusError, FatalError: "invalid job"}
	}

	// Single execution flow (no test cases)
	if len(j.TestCases) == 0 {
		res := s.exec.Run(j.Language, j.Code, j.Input)
		if res.Error != "" {
			status := result.StatusError
			if res.Error == "timeout" {
				status = result.StatusTLE
			}
			return result.Result{
				Status: status,
				Output: res.Output,
				Error:  res.Error,
			}
		}
		return result.Result{
			Status: result.StatusSuccess,
			Output: res.Output,
		}
	}

	// Batch test case execution — compile once, run N times concurrently
	return s.executeBatch(j)
}

func (s *Service) executeBatch(j job.Job) result.Result {
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

	finalResult := result.Result{
		TestCases: make([]result.TestCaseResult, 0, len(j.TestCases)),
		Total:     len(j.TestCases),
	}

	if err != nil {
		return result.Result{
			Status:     result.StatusError,
			FatalError: err.Error(),
		}
	}

	// Evaluate results concurrently and cancel on first failure.
	// Since batchResults are indexed, we evaluate in parallel but build
	// the final result slice in order.
	tcResults := make([]result.TestCaseResult, len(j.TestCases))
	var passed int
	var mu sync.Mutex
	var failed bool
	var wg sync.WaitGroup

	for i, tc := range j.TestCases {
		wg.Add(1)
		go func(idx int, tc job.TestCase) {
			defer wg.Done()

			res := batchResults[idx]
			tcRes := result.TestCaseResult{
				Input:          tc.Input,
				ExpectedOutput: tc.ExpectedOutput,
			}

			if res.Error == "cancelled" {
				tcRes.Status = result.StatusError
				tcRes.Error = "skipped: previous test case failed"
			} else if res.Error != "" {
				if res.Error == "timeout" {
					tcRes.Status = result.StatusTLE
				} else {
					tcRes.Status = result.StatusError
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
					tcRes.Status = result.StatusSuccess
					mu.Lock()
					passed++
					mu.Unlock()
				} else {
					tcRes.Status = result.StatusFailed
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
		finalResult.Status = result.StatusSuccess
	} else {
		finalResult.Status = result.StatusFailed
	}

	return finalResult
}
