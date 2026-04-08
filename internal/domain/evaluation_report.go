package domain

import "time"

// EvaluationDecision represents the outcome of an evaluation.
const (
	EvaluationDecisionAccepted = "Accepted"
	EvaluationDecisionRejected = "Rejected"
)

// EvaluationReport holds the result of evaluating a patch proposal.
type EvaluationReport struct {
	EvaluationId string `json:"evaluation_id"`
	ExecutionId  string `json:"execution_id"`

	TestPassed      bool `json:"test_passed"`
	LintPassed      bool `json:"lint_passed"`
	TypeCheckPassed bool `json:"typecheck_passed"`

	ChangedFileCount int `json:"changed_file_count"`
	ChangedLineCount int `json:"changed_line_count"`

	EvaluationDecision string   `json:"evaluation_decision"`
	FailureReasons     []string `json:"failure_reasons,omitempty"`

	GeneratedAt time.Time `json:"generated_at"`
}
