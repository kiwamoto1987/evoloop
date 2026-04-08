package domain

import "time"

// IssueStatus represents the lifecycle state of an ImplementationIssue.
const (
	IssueStatusOpen     = "Open"
	IssueStatusProposed = "Proposed"
	IssueStatusAccepted = "Accepted"
	IssueStatusRejected = "Rejected"
)

// IssueCategory classifies the type of issue detected.
const (
	IssueCategoryTestFailure      = "test_failure"
	IssueCategoryLintViolation    = "lint_violation"
	IssueCategoryTypeCheckFailure = "typecheck_failure"
	IssueCategoryEnvironment      = "environment"
)

// IsProposable returns true if the issue can be addressed by LLM patch generation.
func (i *ImplementationIssue) IsProposable() bool {
	return i.IssueCategory != IssueCategoryEnvironment
}

// ImplementationIssue represents a self-improvement issue detected by analysis.
type ImplementationIssue struct {
	IssueId            string    `json:"issue_id"`
	IssueTitle         string    `json:"issue_title"`
	IssueDescription   string    `json:"issue_description"`
	IssueCategory      string    `json:"issue_category"`
	IssuePriority      int       `json:"issue_priority"`
	IssueStatus        string    `json:"issue_status"`
	TargetPaths        []string  `json:"target_paths,omitempty"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}
