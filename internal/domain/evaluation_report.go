package domain

import "time"

// EvaluationDecision represents the outcome of an evaluation.
const (
	EvaluationDecisionAccepted = "Accepted"
	EvaluationDecisionRejected = "Rejected"
)

// CheckStatus represents the result of an individual quality check.
const (
	CheckStatusPassed  = "passed"
	CheckStatusFailed  = "failed"
	CheckStatusSkipped = "skipped"
)

// EvaluationReport holds the result of evaluating a patch proposal.
type EvaluationReport struct {
	EvaluationId string `json:"evaluation_id"`
	ExecutionId  string `json:"execution_id"`

	EvaluationMode  string `json:"evaluation_mode"`
	TestStatus      string `json:"test_status"`
	LintStatus      string `json:"lint_status"`
	TypeCheckStatus string `json:"typecheck_status"`
	ValidateStatus  string `json:"validate_status"`

	ChangedFileCount int `json:"changed_file_count"`
	ChangedLineCount int `json:"changed_line_count"`

	EvaluationDecision string   `json:"evaluation_decision"`
	FailureReasons     []string `json:"failure_reasons,omitempty"`

	GeneratedAt time.Time `json:"generated_at"`
}
