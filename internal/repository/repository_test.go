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

func TestExecutionHistoryRepository_FindByID(t *testing.T) {
	tdb := setupTestDB(t)
	issueRepo := repository.NewImplementationIssueRepository(tdb.DB)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId: "ISSUE001", IssueTitle: "t", IssueDescription: "d",
		IssueCategory: "test_failure", IssuePriority: 1, IssueStatus: "Open",
		CreatedAt: time.Now(),
	}
	if err := issueRepo.Save(issue); err != nil {
		t.Fatalf("failed to save issue: %v", err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC001", IssueId: "ISSUE001",
		ExecutionStatus: "Completed", ModelProvider: "claude", ModelName: "sonnet",
		StartedAt: time.Now().Truncate(time.Second), FinishedAt: time.Now().Truncate(time.Second),
	}
	if err := execRepo.Save(record); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	found, err := execRepo.FindByID("EXEC001")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if found.ExecutionStatus != "Completed" {
		t.Errorf("expected Completed, got %q", found.ExecutionStatus)
	}
	if found.ModelProvider != "claude" {
		t.Errorf("expected claude, got %q", found.ModelProvider)
	}
}

func TestExecutionHistoryRepository_FindByID_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)

	_, err := execRepo.FindByID("NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent ID")
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
	if err := issueRepo.Save(issue); err != nil {
		t.Fatalf("failed to save issue: %v", err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC001", IssueId: "ISSUE001",
		ExecutionStatus: "Completed", StartedAt: time.Now(), FinishedAt: time.Now(),
	}
	if err := execRepo.Save(record); err != nil {
		t.Fatalf("failed to save execution: %v", err)
	}

	report := &domain.EvaluationReport{
		EvaluationId:       "EVAL001",
		ExecutionId:        "EXEC001",
		EvaluationMode:     "sandbox",
		TestStatus:         domain.CheckStatusPassed,
		LintStatus:         domain.CheckStatusPassed,
		TypeCheckStatus:    domain.CheckStatusPassed,
		ValidateStatus:     domain.CheckStatusSkipped,
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
	if err := issueRepo.Save(issue); err != nil {
		t.Fatalf("failed to save issue: %v", err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC002", IssueId: "ISSUE002",
		ExecutionStatus: "Completed", StartedAt: time.Now(), FinishedAt: time.Now(),
	}
	if err := execRepo.Save(record); err != nil {
		t.Fatalf("failed to save execution: %v", err)
	}

	report := &domain.EvaluationReport{
		EvaluationId:       "EVAL002",
		ExecutionId:        "EXEC002",
		EvaluationMode:     "sandbox",
		TestStatus:         domain.CheckStatusFailed,
		LintStatus:         domain.CheckStatusFailed,
		TypeCheckStatus:    domain.CheckStatusSkipped,
		ValidateStatus:     domain.CheckStatusSkipped,
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

func TestImprovementMemoryRepository_SaveAndFind(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImprovementMemoryRepository(tdb.DB)

	entry := &domain.ImprovementMemoryEntry{
		MemoryId:           "MEM001",
		PatternKey:         "test_failure",
		PatternDescription: "test_failure",
		SuccessCount:       3,
		FailureCount:       1,
		LastObservedAt:     time.Now().Truncate(time.Second),
	}

	if err := repo.Save(entry); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	found, err := repo.FindByPatternKey("test_failure")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if found.SuccessCount != 3 {
		t.Errorf("expected SuccessCount 3, got %d", found.SuccessCount)
	}
	if found.FailureCount != 1 {
		t.Errorf("expected FailureCount 1, got %d", found.FailureCount)
	}
}

func TestImprovementMemoryRepository_RecordSuccessAndFailure(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImprovementMemoryRepository(tdb.DB)

	entry := &domain.ImprovementMemoryEntry{
		MemoryId:           "MEM002",
		PatternKey:         "lint_violation",
		PatternDescription: "lint_violation",
		SuccessCount:       0,
		FailureCount:       0,
		LastObservedAt:     time.Now(),
	}
	if err := repo.Save(entry); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	if err := repo.RecordSuccess("lint_violation"); err != nil {
		t.Fatalf("failed to record success: %v", err)
	}
	if err := repo.RecordSuccess("lint_violation"); err != nil {
		t.Fatalf("failed to record success: %v", err)
	}
	if err := repo.RecordFailure("lint_violation"); err != nil {
		t.Fatalf("failed to record failure: %v", err)
	}

	found, err := repo.FindByPatternKey("lint_violation")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if found.SuccessCount != 2 {
		t.Errorf("expected SuccessCount 2, got %d", found.SuccessCount)
	}
	if found.FailureCount != 1 {
		t.Errorf("expected FailureCount 1, got %d", found.FailureCount)
	}
}

func TestImprovementMemoryRepository_FindAll(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImprovementMemoryRepository(tdb.DB)

	for _, key := range []string{"test_failure", "lint_violation"} {
		entry := &domain.ImprovementMemoryEntry{
			MemoryId:           "MEM_" + key,
			PatternKey:         key,
			PatternDescription: key,
			LastObservedAt:     time.Now(),
		}
		if err := repo.Save(entry); err != nil {
			t.Fatalf("failed to save: %v", err)
		}
	}

	entries, err := repo.FindAll()
	if err != nil {
		t.Fatalf("failed to find all: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHookExecutionRepository_SaveAndFind(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewHookExecutionRepository(tdb.DB)

	// Need execution record for FK
	issueRepo := repository.NewImplementationIssueRepository(tdb.DB)
	execRepo := repository.NewExecutionHistoryRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId: "ISSUE_HOOK1", IssueTitle: "t", IssueDescription: "d",
		IssueCategory: "test_failure", IssuePriority: 1, IssueStatus: "Open",
		Source: "analyze", RemediationType: "code_patch", CreatedAt: time.Now(),
	}
	if err := issueRepo.Save(issue); err != nil {
		t.Fatalf("failed to save issue: %v", err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC_HOOK1", IssueId: "ISSUE_HOOK1",
		ExecutionStatus: "Completed", StartedAt: time.Now(), FinishedAt: time.Now(),
	}
	if err := execRepo.Save(record); err != nil {
		t.Fatalf("failed to save execution: %v", err)
	}

	hookRecord := &domain.HookExecutionRecord{
		HookId:      "HOOK001",
		ExecutionId: "EXEC_HOOK1",
		HookType:    "post_apply",
		Command:     "systemctl",
		Args:        []string{"restart", "trade-bot"},
		ExitCode:    0,
		Stdout:      "ok",
		Stderr:      "",
		DurationMs:  1500,
		TimedOut:    false,
		ExecutedAt:  time.Now().Truncate(time.Second),
	}

	if err := repo.Save(hookRecord); err != nil {
		t.Fatalf("failed to save hook record: %v", err)
	}

	records, err := repo.FindByExecutionID("EXEC_HOOK1")
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Command != "systemctl" {
		t.Errorf("expected command 'systemctl', got %q", records[0].Command)
	}
	if records[0].ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", records[0].ExitCode)
	}
	if len(records[0].Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(records[0].Args))
	}
}

func TestHookExecutionRepository_FindByExecutionID_Empty(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewHookExecutionRepository(tdb.DB)

	records, err := repo.FindByExecutionID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestIssueRepository_FindOpenProposable(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	issues := []*domain.ImplementationIssue{
		{IssueId: "I1", IssueTitle: "t1", IssueDescription: "d", IssueCategory: domain.IssueCategoryTestFailure, IssuePriority: 2, IssueStatus: domain.IssueStatusOpen, Source: "analyze", RemediationType: "code_patch", CreatedAt: time.Now()},
		{IssueId: "I2", IssueTitle: "t2", IssueDescription: "d", IssueCategory: domain.IssueCategoryEnvironment, IssuePriority: 0, IssueStatus: domain.IssueStatusOpen, Source: "analyze", RemediationType: "code_patch", CreatedAt: time.Now()},
		{IssueId: "I3", IssueTitle: "t3", IssueDescription: "d", IssueCategory: domain.IssueCategoryKPIDegradation, IssuePriority: 1, IssueStatus: domain.IssueStatusOpen, Source: "external", RemediationType: "config_patch", CreatedAt: time.Now()},
		{IssueId: "I4", IssueTitle: "t4", IssueDescription: "d", IssueCategory: domain.IssueCategoryLintViolation, IssuePriority: 1, IssueStatus: domain.IssueStatusRejected, Source: "analyze", RemediationType: "code_patch", CreatedAt: time.Now()},
	}

	for _, issue := range issues {
		if err := repo.Save(issue); err != nil {
			t.Fatalf("failed to save issue: %v", err)
		}
	}

	found, err := repo.FindOpenProposable()
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}

	// Should return I1 and I3 (Open + proposable), NOT I2 (environment) or I4 (Rejected)
	if len(found) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(found))
	}
	// Ordered by priority ASC: I3 (pri=1) before I1 (pri=2)
	if found[0].IssueId != "I3" {
		t.Errorf("expected first issue I3, got %q", found[0].IssueId)
	}
	if found[1].IssueId != "I1" {
		t.Errorf("expected second issue I1, got %q", found[1].IssueId)
	}
}

func TestIssueRepository_FindByDedupKey(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId: "DEDUP1", IssueTitle: "slippage", IssueDescription: "d",
		IssueCategory: domain.IssueCategoryKPIDegradation, IssuePriority: 1,
		IssueStatus: domain.IssueStatusOpen, DedupKey: "kpi:slippage:arb",
		Source: "external", RemediationType: "config_patch", CreatedAt: time.Now(),
	}
	if err := repo.Save(issue); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	found, err := repo.FindByDedupKey("kpi:slippage:arb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find issue, got nil")
	}
	if found.IssueId != "DEDUP1" {
		t.Errorf("expected issue DEDUP1, got %q", found.IssueId)
	}
}

func TestIssueRepository_FindByDedupKey_NotFound(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	found, err := repo.FindByDedupKey("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestIssueRepository_FindByDedupKey_IgnoresNonOpen(t *testing.T) {
	tdb := setupTestDB(t)
	repo := repository.NewImplementationIssueRepository(tdb.DB)

	issue := &domain.ImplementationIssue{
		IssueId: "DEDUP2", IssueTitle: "closed", IssueDescription: "d",
		IssueCategory: domain.IssueCategoryKPIDegradation, IssuePriority: 1,
		IssueStatus: domain.IssueStatusCompleted, DedupKey: "kpi:closed",
		Source: "external", RemediationType: "config_patch", CreatedAt: time.Now(),
	}
	if err := repo.Save(issue); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	found, err := repo.FindByDedupKey("kpi:closed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil for non-Open issue, got %+v", found)
	}
}
