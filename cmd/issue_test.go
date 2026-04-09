package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/cmd"
	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
)

func setupIssueTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	configDir := filepath.Join(dir, ".evoloop")
	if err := os.MkdirAll(filepath.Join(configDir, "runtime"), 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `project_name: test
llm:
  provider: claude
  model: sonnet
  command: "claude"
issues:
  allowed_categories:
    - "kpi_degradation"
    - "config_tuning"
    - "test_failure"
  max_priority: 10
  max_description_length: 5000
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestIssueCreate_Success(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "High slippage",
		"--description", "Avg slippage 2.1%",
		"--category", "kpi_degradation",
		"--remediation", "config_patch",
		"--priority", "1",
		"--source", "check_kpi.sh",
		"--source-ref", "rule:slippage",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify issue in DB
	db, err := repository.OpenDatabase(config.DatabasePath(dir))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := repository.NewImplementationIssueRepository(db)
	issues, err := repo.FindAll()
	if err != nil {
		t.Fatalf("failed to find issues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueTitle != "High slippage" {
		t.Errorf("expected title 'High slippage', got %q", issues[0].IssueTitle)
	}
	if issues[0].Source != "check_kpi.sh" {
		t.Errorf("expected source 'check_kpi.sh', got %q", issues[0].Source)
	}
	if issues[0].RemediationType != domain.RemediationTypeConfigPatch {
		t.Errorf("expected remediation 'config_patch', got %q", issues[0].RemediationType)
	}
}

func TestIssueCreate_EmptyTitle(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "",
		"--description", "test",
		"--category", "kpi_degradation",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestIssueCreate_InvalidCategory(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "test",
		"--description", "test",
		"--category", "invalid_category",
	})
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestIssueCreate_InvalidPriority(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "test",
		"--description", "test",
		"--category", "kpi_degradation",
		"--priority", "99",
	})
	if err == nil {
		t.Fatal("expected error for priority exceeding max")
	}
}

func TestIssueCreate_Dedup_CreatesThenUpdates(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// First create
	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "slippage",
		"--description", "slippage 2.1%",
		"--category", "kpi_degradation",
		"--priority", "2",
		"--dedup-key", "kpi:slippage:arb",
	})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Second create with same dedup key = update
	err = cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "slippage updated",
		"--description", "slippage 3.0%",
		"--category", "kpi_degradation",
		"--priority", "1",
		"--dedup-key", "kpi:slippage:arb",
	})
	if err != nil {
		t.Fatalf("second create (dedup) failed: %v", err)
	}

	// Should still be 1 issue, with updated description
	db, err := repository.OpenDatabase(config.DatabasePath(dir))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := repository.NewImplementationIssueRepository(db)
	issues, err := repo.FindAll()
	if err != nil {
		t.Fatalf("failed to find issues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue (dedup), got %d", len(issues))
	}
	if issues[0].IssueDescription != "slippage 3.0%" {
		t.Errorf("expected updated description, got %q", issues[0].IssueDescription)
	}
	if issues[0].IssuePriority != 1 {
		t.Errorf("expected updated priority 1, got %d", issues[0].IssuePriority)
	}
}

func TestIssueCreate_InvalidRemediation(t *testing.T) {
	dir := setupIssueTestProject(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := cmd.ExecuteArgs([]string{
		"issue", "create",
		"--title", "test",
		"--description", "test",
		"--category", "kpi_degradation",
		"--remediation", "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid remediation type")
	}
}
