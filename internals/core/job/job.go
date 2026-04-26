package job

type TestCase struct {
	Input          string
	ExpectedOutput string
}

type Job struct {
	Code      string
	Language  string
	Input     string
	TestCases []TestCase
}
