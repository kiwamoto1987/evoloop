package service_test

import (
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestIssueSelector_SelectNext_BasicPriority(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 3, CooldownMinutes: 60})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryLintViolation, IssuePriority: 3, IssueStatus: domain.IssueStatusOpen},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen},
		{IssueId: "C", IssueCategory: domain.IssueCategoryKPIDegradation, IssuePriority: 2, IssueStatus: domain.IssueStatusOpen},
	}

	selected := sel.SelectNext(issues)
	if selected == nil {
		t.Fatal("expected an issue, got nil")
	}
	if selected.IssueId != "B" {
		t.Errorf("expected B (lowest priority), got %q", selected.IssueId)
	}
}

func TestIssueSelector_SelectNext_FiltersEnvironment(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 3, CooldownMinutes: 60})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryEnvironment, IssuePriority: 0, IssueStatus: domain.IssueStatusOpen},
	}

	selected := sel.SelectNext(issues)
	if selected != nil {
		t.Errorf("expected nil (environment not proposable), got %q", selected.IssueId)
	}
}

func TestIssueSelector_SelectNext_FiltersExceededMaxAttempts(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 3, CooldownMinutes: 0})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 3},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 2, IssueStatus: domain.IssueStatusOpen, AttemptCount: 1},
	}

	selected := sel.SelectNext(issues)
	if selected == nil {
		t.Fatal("expected an issue, got nil")
	}
	if selected.IssueId != "B" {
		t.Errorf("expected B (A exceeded max attempts), got %q", selected.IssueId)
	}
}

func TestIssueSelector_SelectNext_FiltersCooldown(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 3, CooldownMinutes: 60})

	now := time.Now()
	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 1, LastAttemptedAt: now.Add(-30 * time.Minute)},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 2, IssueStatus: domain.IssueStatusOpen, AttemptCount: 1, LastAttemptedAt: now.Add(-90 * time.Minute)},
	}

	selected := sel.SelectNext(issues)
	if selected == nil {
		t.Fatal("expected an issue, got nil")
	}
	if selected.IssueId != "B" {
		t.Errorf("expected B (A still in cooldown), got %q", selected.IssueId)
	}
}

func TestIssueSelector_SelectNext_CooldownSkippedForZeroAttempts(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 3, CooldownMinutes: 60})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 0},
	}

	selected := sel.SelectNext(issues)
	if selected == nil {
		t.Fatal("expected an issue (zero attempts, no cooldown), got nil")
	}
}

func TestIssueSelector_SelectNext_PrefersLowerAttemptCount(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 5, CooldownMinutes: 0})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 3},
		{IssueId: "B", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 1},
	}

	selected := sel.SelectNext(issues)
	if selected == nil {
		t.Fatal("expected an issue, got nil")
	}
	if selected.IssueId != "B" {
		t.Errorf("expected B (lower attempt count at same priority), got %q", selected.IssueId)
	}
}

func TestIssueSelector_SelectNext_EmptyList(t *testing.T) {
	sel := service.NewIssueSelector(policy.DefaultPolicy())
	selected := sel.SelectNext(nil)
	if selected != nil {
		t.Errorf("expected nil for empty list, got %+v", selected)
	}
}

func TestIssueSelector_SelectNext_AllFiltered(t *testing.T) {
	sel := service.NewIssueSelector(&policy.ExecutionPolicy{MaxAttempts: 1, CooldownMinutes: 0})

	issues := []*domain.ImplementationIssue{
		{IssueId: "A", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, AttemptCount: 1},
		{IssueId: "B", IssueCategory: domain.IssueCategoryEnvironment, IssuePriority: 0, IssueStatus: domain.IssueStatusOpen},
	}

	selected := sel.SelectNext(issues)
	if selected != nil {
		t.Errorf("expected nil (all filtered), got %q", selected.IssueId)
	}
}
