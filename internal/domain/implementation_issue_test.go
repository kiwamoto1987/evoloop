package domain_test

import (
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

func TestIsProposable_ExistingCategories(t *testing.T) {
	tests := []struct {
		category string
		want     bool
	}{
		{domain.IssueCategoryTestFailure, true},
		{domain.IssueCategoryLintViolation, true},
		{domain.IssueCategoryTypeCheckFailure, true},
		{domain.IssueCategoryEnvironment, false},
		{domain.IssueCategoryKPIDegradation, true},
		{domain.IssueCategoryConfigTuning, true},
	}

	for _, tt := range tests {
		issue := &domain.ImplementationIssue{IssueCategory: tt.category}
		if got := issue.IsProposable(); got != tt.want {
			t.Errorf("IsProposable(%q) = %v, want %v", tt.category, got, tt.want)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{domain.IssueStatusOpen, false},
		{domain.IssueStatusProposed, false},
		{domain.IssueStatusAccepted, false},
		{domain.IssueStatusApplied, false},
		{domain.IssueStatusCompleted, false},
		{domain.IssueStatusRejected, true},
		{domain.IssueStatusApplyFailed, true},
		{domain.IssueStatusHookFailed, true},
		{domain.IssueStatusClosed, false},
	}

	for _, tt := range tests {
		issue := &domain.ImplementationIssue{IssueStatus: tt.status}
		if got := issue.IsRetryable(); got != tt.want {
			t.Errorf("IsRetryable(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestIssueStatusConstants(t *testing.T) {
	statuses := []string{
		domain.IssueStatusOpen,
		domain.IssueStatusProposed,
		domain.IssueStatusAccepted,
		domain.IssueStatusApplied,
		domain.IssueStatusCompleted,
		domain.IssueStatusRejected,
		domain.IssueStatusApplyFailed,
		domain.IssueStatusHookFailed,
		domain.IssueStatusClosed,
	}

	seen := make(map[string]bool)
	for _, s := range statuses {
		if s == "" {
			t.Error("status constant should not be empty")
		}
		if seen[s] {
			t.Errorf("duplicate status constant: %q", s)
		}
		seen[s] = true
	}
}

func TestNewCategoryConstants(t *testing.T) {
	if domain.IssueCategoryKPIDegradation == "" {
		t.Error("IssueCategoryKPIDegradation should not be empty")
	}
	if domain.IssueCategoryConfigTuning == "" {
		t.Error("IssueCategoryConfigTuning should not be empty")
	}
}

func TestRemediationTypeConstants(t *testing.T) {
	if domain.RemediationTypeCodePatch == "" {
		t.Error("RemediationTypeCodePatch should not be empty")
	}
	if domain.RemediationTypeConfigPatch == "" {
		t.Error("RemediationTypeConfigPatch should not be empty")
	}
}

func TestIssueSourceConstants(t *testing.T) {
	if domain.IssueSourceAnalyze == "" {
		t.Error("IssueSourceAnalyze should not be empty")
	}
	if domain.IssueSourceExternal == "" {
		t.Error("IssueSourceExternal should not be empty")
	}
}

func TestNewFieldsExist(t *testing.T) {
	now := time.Now()
	issue := &domain.ImplementationIssue{
		IssueId:          "test-id",
		IssueTitle:       "test",
		IssueDescription: "desc",
		IssueCategory:    domain.IssueCategoryKPIDegradation,
		RemediationType:  domain.RemediationTypeConfigPatch,
		IssuePriority:    1,
		IssueStatus:      domain.IssueStatusOpen,
		Source:           domain.IssueSourceExternal,
		SourceRef:        "rule:slippage",
		DedupKey:         "kpi:slippage:arb",
		AttemptCount:     0,
		LastAttemptedAt:  now,
		CreatedAt:        now,
	}

	if issue.RemediationType != domain.RemediationTypeConfigPatch {
		t.Errorf("RemediationType = %q, want %q", issue.RemediationType, domain.RemediationTypeConfigPatch)
	}
	if issue.Source != domain.IssueSourceExternal {
		t.Errorf("Source = %q, want %q", issue.Source, domain.IssueSourceExternal)
	}
	if issue.DedupKey != "kpi:slippage:arb" {
		t.Errorf("DedupKey = %q, want %q", issue.DedupKey, "kpi:slippage:arb")
	}
	if issue.AttemptCount != 0 {
		t.Errorf("AttemptCount = %d, want 0", issue.AttemptCount)
	}
}
