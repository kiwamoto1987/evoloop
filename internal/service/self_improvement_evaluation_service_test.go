package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func setupEvalProject(t *testing.T) (string, string) {
	t.Helper()

	// Create a project directory with a simple file
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a patch that modifies the file
	patchContent := `diff --git a/hello.txt b/hello.txt
--- a/hello.txt
+++ b/hello.txt
@@ -1 +1 @@
-hello
+hello world
`
	patchDir := t.TempDir()
	patchPath := filepath.Join(patchDir, "test.patch")
	if err := os.WriteFile(patchPath, []byte(patchContent), 0644); err != nil {
		t.Fatal(err)
	}

	return projectDir, patchPath
}

func TestEvaluate_Accepted(t *testing.T) {
	projectDir, patchPath := setupEvalProject(t)

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC001",
		PatchPath:   patchPath,
	}

	projectCtx := &domain.ProjectContext{
		ProjectRootPath: projectDir,
		// No commands configured = all pass
	}

	svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
	report, err := svc.Evaluate(record, projectCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.EvaluationDecision != domain.EvaluationDecisionAccepted {
		t.Errorf("expected Accepted, got %q (reasons: %v)", report.EvaluationDecision, report.FailureReasons)
	}
	if report.EvaluationId == "" {
		t.Error("expected EvaluationId to be set")
	}
	if report.ExecutionId != "EXEC001" {
		t.Errorf("expected ExecutionId 'EXEC001', got %q", report.ExecutionId)
	}
}

func TestEvaluate_RejectedByTestFailure(t *testing.T) {
	projectDir, patchPath := setupEvalProject(t)

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC002",
		PatchPath:   patchPath,
	}

	projectCtx := &domain.ProjectContext{
		ProjectRootPath: projectDir,
		TestCommand:     "false", // always fails
	}

	svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
	report, err := svc.Evaluate(record, projectCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.EvaluationDecision != domain.EvaluationDecisionRejected {
		t.Errorf("expected Rejected, got %q", report.EvaluationDecision)
	}
	if report.TestPassed {
		t.Error("expected TestPassed to be false")
	}
}

func TestEvaluate_RejectedByPolicy(t *testing.T) {
	projectDir, patchPath := setupEvalProject(t)

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC003",
		PatchPath:   patchPath,
	}

	projectCtx := &domain.ProjectContext{
		ProjectRootPath: projectDir,
	}

	// Set very restrictive policy
	restrictivePolicy := &policy.ExecutionPolicy{
		MaxFiles: 0,
		MaxLines: 0,
	}

	svc := service.NewSelfImprovementEvaluationService(restrictivePolicy)
	report, err := svc.Evaluate(record, projectCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.EvaluationDecision != domain.EvaluationDecisionRejected {
		t.Errorf("expected Rejected, got %q", report.EvaluationDecision)
	}
	if len(report.FailureReasons) == 0 {
		t.Error("expected failure reasons to be set")
	}
}

func TestEvaluate_InvalidPatchPath(t *testing.T) {
	projectDir := t.TempDir()

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC004",
		PatchPath:   "/nonexistent/patch.patch",
	}

	projectCtx := &domain.ProjectContext{
		ProjectRootPath: projectDir,
	}

	svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
	_, err := svc.Evaluate(record, projectCtx)
	if err == nil {
		t.Fatal("expected error for invalid patch path")
	}
}

func TestEvaluate_CopyExcludesPatterns(t *testing.T) {
	projectDir := t.TempDir()

	// Create directories that should be excluded
	for _, dir := range []string{".git", "node_modules", ".evoloop"} {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(projectDir, dir, "test.txt"), []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a valid file and patch
	if err := os.WriteFile(filepath.Join(projectDir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patchContent := `diff --git a/hello.txt b/hello.txt
--- a/hello.txt
+++ b/hello.txt
@@ -1 +1 @@
-hello
+hello world
`
	patchDir := t.TempDir()
	patchPath := filepath.Join(patchDir, "test.patch")
	if err := os.WriteFile(patchPath, []byte(patchContent), 0644); err != nil {
		t.Fatal(err)
	}

	record := &domain.ExecutionRecord{
		ExecutionId: "EXEC005",
		PatchPath:   patchPath,
	}

	projectCtx := &domain.ProjectContext{
		ProjectRootPath: projectDir,
	}

	svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
	report, err := svc.Evaluate(record, projectCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should succeed since excluded dirs don't interfere
	if report.EvaluationDecision != domain.EvaluationDecisionAccepted {
		t.Errorf("expected Accepted, got %q (reasons: %v)", report.EvaluationDecision, report.FailureReasons)
	}
}
