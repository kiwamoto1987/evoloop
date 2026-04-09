package domain

import "time"

// IssueStatus represents the lifecycle state of an ImplementationIssue.
const (
	IssueStatusOpen        = "Open"
	IssueStatusProposed    = "Proposed"
	IssueStatusAccepted    = "Accepted"
	IssueStatusApplied     = "Applied"
	IssueStatusCompleted   = "Completed"
	IssueStatusRejected    = "Rejected"
	IssueStatusApplyFailed = "ApplyFailed"
	IssueStatusHookFailed  = "HookFailed"
	IssueStatusClosed      = "Closed"
)

// IssueCategory classifies the type of issue detected.
const (
	IssueCategoryTestFailure      = "test_failure"
	IssueCategoryLintViolation    = "lint_violation"
	IssueCategoryTypeCheckFailure = "typecheck_failure"
	IssueCategoryEnvironment      = "environment"
	IssueCategoryKPIDegradation   = "kpi_degradation"
	IssueCategoryConfigTuning     = "config_tuning"
)

// RemediationType describes how an issue should be remediated.
const (
	RemediationTypeCodePatch   = "code_patch"
	RemediationTypeConfigPatch = "config_patch"
)

// IssueSource identifies where the issue originated.
const (
	IssueSourceAnalyze  = "analyze"
	IssueSourceExternal = "external"
)

// IsProposable returns true if the issue can be addressed by LLM patch generation.
func (i *ImplementationIssue) IsProposable() bool {
	return i.IssueCategory != IssueCategoryEnvironment
}

// IsRetryable returns true if the issue is in a failed state that allows retry.
func (i *ImplementationIssue) IsRetryable() bool {
	return i.IssueStatus == IssueStatusRejected ||
		i.IssueStatus == IssueStatusApplyFailed ||
		i.IssueStatus == IssueStatusHookFailed
}

// ImplementationIssue represents a self-improvement issue detected by analysis.
type ImplementationIssue struct {
	IssueId            string    `json:"issue_id"`
	IssueTitle         string    `json:"issue_title"`
	IssueDescription   string    `json:"issue_description"`
	IssueCategory      string    `json:"issue_category"`
	RemediationType    string    `json:"remediation_type"`
	IssuePriority      int       `json:"issue_priority"`
	IssueStatus        string    `json:"issue_status"`
	TargetPaths        []string  `json:"target_paths,omitempty"`
	AcceptanceCriteria []string  `json:"acceptance_criteria,omitempty"`
	Source             string    `json:"source"`
	SourceRef          string    `json:"source_ref,omitempty"`
	DedupKey           string    `json:"dedup_key,omitempty"`
	AttemptCount       int       `json:"attempt_count"`
	LastAttemptedAt    time.Time `json:"last_attempted_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}
