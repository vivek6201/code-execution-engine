package evaluator

import (
	"regexp"
	"strings"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Evaluate(actualOutput string, expectedOutput string) bool {
	actualOutput = clean(actualOutput)
	expectedOutput = clean(expectedOutput)

	return actualOutput == expectedOutput
}

func clean(s string) string {
	// Remove leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Remove multiple spaces (convert to single space)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	// Convert to lowercase (optional, but good for robust comparison)
	s = strings.ToLower(s)

	return s
}
