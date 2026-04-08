package cmd

import (
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

func TestSelectBestIssue_HighestPriority(t *testing.T) {
	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryLintViolation, IssuePriority: 3},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1},
		{IssueId: "C", IssueCategory: domain.IssueCategoryTypeCheckFailure, IssuePriority: 2},
	}

	best := selectBestIssue(issues)
	if best == nil {
		t.Fatal("expected an issue to be selected")
	}
	if best.IssueId != "B" {
		t.Errorf("expected issue B (priority 1), got %s (priority %d)", best.IssueId, best.IssuePriority)
	}
}

func TestSelectBestIssue_SkipsEnvironmentIssues(t *testing.T) {
	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryEnvironment, IssuePriority: 0},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1},
	}

	best := selectBestIssue(issues)
	if best == nil {
		t.Fatal("expected an issue to be selected")
	}
	if best.IssueId != "B" {
		t.Errorf("expected issue B, got %s", best.IssueId)
	}
}

func TestSelectBestIssue_AllEnvironment(t *testing.T) {
	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryEnvironment, IssuePriority: 0},
	}

	best := selectBestIssue(issues)
	if best != nil {
		t.Error("expected nil when all issues are environment")
	}
}

func TestSelectBestIssue_Empty(t *testing.T) {
	best := selectBestIssue(nil)
	if best != nil {
		t.Error("expected nil for empty list")
	}
}
