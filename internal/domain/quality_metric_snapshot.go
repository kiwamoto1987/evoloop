package domain

// QualityMetricSnapshot holds the results of running quality checks.
type QualityMetricSnapshot struct {
	TestSucceeded      bool   `json:"test_succeeded"`
	LintSucceeded      bool   `json:"lint_succeeded"`
	TypeCheckSucceeded bool   `json:"typecheck_succeeded"`
	TestOutput         string `json:"test_output,omitempty"`
	LintOutput         string `json:"lint_output,omitempty"`
	TypeCheckOutput    string `json:"typecheck_output,omitempty"`

	TestToolMissing      bool `json:"test_tool_missing,omitempty"`
	LintToolMissing      bool `json:"lint_tool_missing,omitempty"`
	TypeCheckToolMissing bool `json:"typecheck_tool_missing,omitempty"`
}
