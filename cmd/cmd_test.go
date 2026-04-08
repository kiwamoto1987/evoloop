package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/cmd"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init git repo
	run := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %s\n%s", args, err, out)
		}
	}

	run("init")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial")

	return dir
}

func TestInspectCommand_NonGitDir(t *testing.T) {
	dir := t.TempDir()
	err := cmd.ExecuteArgs([]string{"inspect", "--path", dir})
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestInspectCommand_ValidGitDir(t *testing.T) {
	dir := setupTestProject(t)
	err := cmd.ExecuteArgs([]string{"inspect", "--path", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeCommand_NonGitDir(t *testing.T) {
	dir := t.TempDir()

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	err = cmd.ExecuteArgs([]string{"analyze"})
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestProposeCommand_MissingIssueFlag(t *testing.T) {
	err := cmd.ExecuteArgs([]string{"propose"})
	if err == nil {
		t.Fatal("expected error for missing --issue flag")
	}
}

func TestProposeCommand_IssueNotFound(t *testing.T) {
	dir := setupTestProject(t)

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	// Init evoloop so DB exists
	if err := cmd.RunInit(); err != nil {
		t.Fatal(err)
	}

	err = cmd.ExecuteArgs([]string{"propose", "--issue", "NONEXISTENT"})
	if err == nil {
		t.Fatal("expected error for non-existent issue")
	}
}

func TestEvaluateCommand_MissingExecutionFlag(t *testing.T) {
	err := cmd.ExecuteArgs([]string{"evaluate"})
	if err == nil {
		t.Fatal("expected error for missing --execution flag")
	}
}

func TestEvaluateCommand_ExecutionNotFound(t *testing.T) {
	dir := setupTestProject(t)

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	if err := cmd.RunInit(); err != nil {
		t.Fatal(err)
	}

	err = cmd.ExecuteArgs([]string{"evaluate", "--execution", "NONEXISTENT"})
	if err == nil {
		t.Fatal("expected error for non-existent execution")
	}
}

func TestHistoryCommand_EmptyDB(t *testing.T) {
	dir := setupTestProject(t)

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	if err := cmd.RunInit(); err != nil {
		t.Fatal(err)
	}

	err = cmd.ExecuteArgs([]string{"history"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
