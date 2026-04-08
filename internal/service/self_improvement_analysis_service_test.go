package service_test

import (
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestAnalyze_AllPassed(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      true,
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService(nil)
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

	analyzer := service.NewSelfImprovementAnalysisService(nil)
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

	analyzer := service.NewSelfImprovementAnalysisService(nil)
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

	analyzer := service.NewSelfImprovementAnalysisService(nil)
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

	analyzer := service.NewSelfImprovementAnalysisService(nil)
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

func TestAnalyze_ToolMissing_EnvironmentIssue(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      false,
		LintToolMissing:    true,
		LintOutput:         "tool not found: golangci-lint",
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService(nil)
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueCategory != domain.IssueCategoryEnvironment {
		t.Errorf("expected category %q, got %q", domain.IssueCategoryEnvironment, issues[0].IssueCategory)
	}
	if issues[0].IsProposable() {
		t.Error("environment issue should not be proposable")
	}
}

func TestAnalyze_ToolExists_QualityIssue(t *testing.T) {
	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      false,
		TestToolMissing:    false,
		TestOutput:         "FAIL: TestSomething",
		LintSucceeded:      true,
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService(nil)
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueCategory != domain.IssueCategoryTestFailure {
		t.Errorf("expected category %q, got %q", domain.IssueCategoryTestFailure, issues[0].IssueCategory)
	}
	if !issues[0].IsProposable() {
		t.Error("code quality issue should be proposable")
	}
}

func TestAnalyze_PriorityBoostFromMemory(t *testing.T) {
	tdb, err := repository.OpenTestDatabase()
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tdb.Close()

	memoryRepo := repository.NewImprovementMemoryRepository(tdb.DB)

	// Create a memory entry with high success rate
	entry := &domain.ImprovementMemoryEntry{
		MemoryId:           "MEM001",
		PatternKey:         domain.IssueCategoryLintViolation,
		PatternDescription: "lint_violation",
		SuccessCount:       8,
		FailureCount:       2,
		LastObservedAt:     time.Now(),
	}
	if err := memoryRepo.Save(entry); err != nil {
		t.Fatalf("failed to save memory: %v", err)
	}

	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      true,
		LintSucceeded:      false,
		LintOutput:         "lint error",
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService(memoryRepo)
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	// High success rate (80%) should boost priority from 2 to 1
	if issues[0].IssuePriority != 1 {
		t.Errorf("expected boosted priority 1, got %d", issues[0].IssuePriority)
	}
}

func TestAnalyze_PriorityDeprioritizeFromMemory(t *testing.T) {
	tdb, err := repository.OpenTestDatabase()
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer tdb.Close()

	memoryRepo := repository.NewImprovementMemoryRepository(tdb.DB)

	// Create a memory entry with high failure rate
	entry := &domain.ImprovementMemoryEntry{
		MemoryId:           "MEM002",
		PatternKey:         domain.IssueCategoryTestFailure,
		PatternDescription: "test_failure",
		SuccessCount:       1,
		FailureCount:       9,
		LastObservedAt:     time.Now(),
	}
	if err := memoryRepo.Save(entry); err != nil {
		t.Fatalf("failed to save memory: %v", err)
	}

	snapshot := &domain.QualityMetricSnapshot{
		TestSucceeded:      false,
		TestOutput:         "test fail",
		LintSucceeded:      true,
		TypeCheckSucceeded: true,
	}

	analyzer := service.NewSelfImprovementAnalysisService(memoryRepo)
	issues := analyzer.Analyze(snapshot)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	// High failure rate (90%) should deprioritize from 1 to 3
	if issues[0].IssuePriority != 3 {
		t.Errorf("expected deprioritized priority 3, got %d", issues[0].IssuePriority)
	}
}
