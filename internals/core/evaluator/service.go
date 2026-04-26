package evaluator

import (
	"strings"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Evaluate(actualOutput string, expectedOutput string) bool {
	return strings.TrimSpace(actualOutput) == strings.TrimSpace(expectedOutput)
}
