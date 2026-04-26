package execution

import (
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

	// Test cases evaluation flow
	finalResult := result.Result{
		TestCases: make([]result.TestCaseResult, 0, len(j.TestCases)),
		Total:     len(j.TestCases),
	}

	for _, tc := range j.TestCases {
		res := s.exec.Run(j.Language, j.Code, tc.Input)

		tcRes := result.TestCaseResult{
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		}

		if res.Error != "" {
			if res.Error == "timeout" {
				tcRes.Status = result.StatusTLE
			} else {
				tcRes.Status = result.StatusError
				tcRes.Error = res.Error
			}
		} else {
			tcRes.ActualOutput = res.Output
			if s.eval.Evaluate(res.Output, tc.ExpectedOutput) {
				tcRes.Status = result.StatusSuccess
				finalResult.Passed++
			} else {
				tcRes.Status = result.StatusFailed
			}
		}
		finalResult.TestCases = append(finalResult.TestCases, tcRes)
	}

	if finalResult.Passed == finalResult.Total {
		finalResult.Status = result.StatusSuccess
	} else {
		finalResult.Status = result.StatusFailed
	}

	return finalResult
}
