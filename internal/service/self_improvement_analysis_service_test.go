package service_test

import (
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestAnalyze_AllPassed(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      true,
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService()
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestAnalyze_TestFailure(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      false,
		TestOutput:         "FAIL: TestSomething",
		LintSucceeded:      true,
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService()
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if issue.IssueCategory != domain.IssueCategoryTestFailure {
		t.Errorf("expected category %q, got %q", domain.IssueCategoryTestFailure, issue.IssueCategory)
	}
	if issue.IssueStatus != domain.IssueStatusOpen {
		t.Errorf("expected status %q, got %q", domain.IssueStatusOpen, issue.IssueStatus)
	}
	if issue.IssuePriority != 1 {
		t.Errorf("expected priority 1, got %d", issue.IssuePriority)
	}
	if issue.IssueId == "" {
		t.Error("expected IssueId to be set")
	}
}

func TestAnalyze_LintFailure(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      false,
		LintOutput:         "unused variable",
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService()
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueCategory != domain.IssueCategoryLintViolation {
		t.Errorf("expected category %q, got %q", domain.IssueCategoryLintViolation, issues[0].IssueCategory)
	}
	if issues[0].IssuePriority != 2 {
		t.Errorf("expected priority 2, got %d", issues[0].IssuePriority)
	}
}

func TestAnalyze_TypeCheckFailure(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      true,
		TypeCheckSucceeded: false,
		TypeCheckOutput:    "type error",
	}

	analyzer := service.NewSelfImprovementAnalysisService()
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueCategory != domain.IssueCategoryTypeCheckFailure {
		t.Errorf("expected category %q, got %q", domain.IssueCategoryTypeCheckFailure, issues[0].IssueCategory)
	}
}

func TestAnalyze_MultipleFailures(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      false,
		TestOutput:         "test fail",
		LintSucceeded:      false,
		LintOutput:         "lint fail",
		TypeCheckSucceeded: false,
		TypeCheckOutput:    "type fail",
	}

	analyzer := service.NewSelfImprovementAnalysisService()
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}

	// Verify each issue has a unique ID
	ids := make(map[string]bool)
	for _, issue := range issues {
		if ids[issue.IssueId] {
			t.Errorf("duplicate issue ID: %s", issue.IssueId)
		}
		ids[issue.IssueId] = true
	}
}
