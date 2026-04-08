package cmd_test

import (
	"os"
	"testing"

	"github.com/kiwamoto1987/evoloop/cmd"
)

func TestRunCommand_NoConfig(t *testing.T) {
	dir := setupTestProject(t)
	chdirTest(t, dir)

	err := cmd.ExecuteArgs([]string{"run"})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}

func TestRunCommand_NoProposableIssues(t *testing.T) {
	dir := setupTestProject(t)
	chdirTest(t, dir)

	if err := cmd.RunInit(); err != nil {
		t.Fatal(err)
	}

	err := cmd.ExecuteArgs([]string{"run", "--max-iterations", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func chdirTest(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
}
