package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
)

func setupTestDB(t *testing.T) *repository.TestDB {
	t.Helper()
	tdb, err := repository.OpenTestDatabase()
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { tdb.Close() })
	return tdb
}

func TestIssueRepository_SaveAndFindByID(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId:            "ISSUE001",
		IssueTitle:         "Fix test failure",
		IssueDescription:   "Tests are failing",
		IssueCategory:      domain.IssueCategoryTestFailure,
		IssuePriority:      1,
		IssueStatus:        domain.IssueStatusOpen,
		TargetPaths:        []string{"internal/service"},
		AcceptanceCriteria: []string{"Tests pass"},
		CreatedAt:          time.Now().Truncate(time.Second),
	}

	if err := repo.Save(issue); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	found, err := repo.FindByID("ISSUE001")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}

	if found.IssueTitle != "Fix test failure" {
		t.Errorf("expected title 'Fix test failure', got %q", found.IssueTitle)
	}
	if found.IssueStatus != domain.IssueStatusOpen {
		t.Errorf("expected status Open, got %q", found.IssueStatus)
	}
}

func TestIssueRepository_FindAll(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	for i, title := range []string{"Issue A", "Issue B"} {
		issue := &domain.ImplementationIssue{
			IssueId:          fmt.Sprintf("ISSUE%03d", i),
			IssueTitle:       title,
			IssueDescription: "desc",
			IssueCategory:    domain.IssueCategoryLintViolation,
			IssuePriority:    2,
			IssueStatus:      domain.IssueStatusOpen,
			CreatedAt:        time.Now().Truncate(time.Second),
		}
		if err := repo.Save(issue); err != nil {
			t.Fatalf("failed to save: %v", err)
		}
	}

	issues, err := repo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestExecutionHistoryRepository_SaveAndFindAll(t *testing.T) {
	tdb := setupTestDB(t)
	issueRepo := repository.NewImplementationIssueRepository(tdb.DB)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)

	// Save prerequisite issue
	issue := &domain.ImplementationIssue{
		IssueId:          "ISSUE001",
		IssueTitle:       "Test issue",
		IssueDescription: "desc",
		IssueCategory:    domain.IssueCategoryTestFailure,
		IssuePriority:    1,
		IssueStatus:      domain.IssueStatusOpen,
		CreatedAt:        time.Now(),
	}
	if err := issueRepo.Save(issue); err != nil {
		t.Fatalf("failed to save issue: %v", err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId:     "EXEC001",
		IssueId:         "ISSUE001",
		ExecutionStatus: domain.ExecutionStatusCompleted,
		ModelProvider:   "claude",
		ModelName:       "sonnet",
		PromptPath:      "/tmp/prompt.txt",
		PatchPath:       "/tmp/patch.patch",
		StartedAt:       time.Now().Truncate(time.Second),
		FinishedAt:      time.Now().Truncate(time.Second),
	}

	if err := execRepo.Save(record); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	records, err := execRepo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].ExecutionStatus != domain.ExecutionStatusCompleted {
		t.Errorf("expected Completed, got %q", records[0].ExecutionStatus)
	}
}

func TestEvaluationReportRepository_SaveAndFindAll(t *testing.T) {
	tdb := setupTestDB(t)
	issueRepo := repository.NewImplementationIssueRepository(tdb.DB)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)
	evalRepo := repository.NewEvaluationReportRepository(tdb.DB)

	// Save prerequisites
	issue := &domain.ImplementationIssue{
		IssueId: "ISSUE001", IssueTitle: "t", IssueDescription: "d",
		IssueCategory: "test_failure", IssuePriority: 1, IssueStatus: "Open",
		CreatedAt: time.Now(),
	}
	issueRepo.Save(issue)

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC001", IssueId: "ISSUE001",
		ExecutionStatus: "Completed", StartedAt: time.Now(), FinishedAt: time.Now(),
	}
	execRepo.Save(record)

	report := &domain.EvaluationReport{
		EvaluationId:       "EVAL001",
		ExecutionId:        "EXEC001",
		TestPassed:         true,
		LintPassed:         true,
		TypeCheckPassed:    true,
		ChangedFileCount:   2,
		ChangedLineCount:   50,
		EvaluationDecision: domain.EvaluationDecisionAccepted,
		GeneratedAt:        time.Now().Truncate(time.Second),
	}

	if err := evalRepo.Save(report); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	reports, err := evalRepo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].EvaluationDecision != domain.EvaluationDecisionAccepted {
		t.Errorf("expected Accepted, got %q", reports[0].EvaluationDecision)
	}
}

func TestEvaluationReportRepository_FailureReasons(t *testing.T) {
	tdb := setupTestDB(t)
	issueRepo := repository.NewImplementationIssueRepository(tdb.DB)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)
	evalRepo := repository.NewEvaluationReportRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId: "ISSUE002", IssueTitle: "t", IssueDescription: "d",
		IssueCategory: "test_failure", IssuePriority: 1, IssueStatus: "Open",
		CreatedAt: time.Now(),
	}
	issueRepo.Save(issue)

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC002", IssueId: "ISSUE002",
		ExecutionStatus: "Completed", StartedAt: time.Now(), FinishedAt: time.Now(),
	}
	execRepo.Save(record)

	report := &domain.EvaluationReport{
		EvaluationId:       "EVAL002",
		ExecutionId:        "EXEC002",
		EvaluationDecision: domain.EvaluationDecisionRejected,
		FailureReasons:     []string{"tests failed", "lint failed"},
		GeneratedAt:        time.Now(),
	}

	if err := evalRepo.Save(report); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	reports, err := evalRepo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}

	if len(reports[0].FailureReasons) != 2 {
		t.Errorf("expected 2 failure reasons, got %d", len(reports[0].FailureReasons))
	}
}
