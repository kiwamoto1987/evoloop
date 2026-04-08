package service_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/service"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %s\n%s", args, err, out)
		}
	}

	run("init")
	run("checkout", "-b", "main")

	// Create a file and commit so HEAD exists
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")

	return dir
}

func TestInspect_GitRepository(t *testing.T) {
	dir := setupGitRepo(t)

	svc := service.NewProjectInspectionService()
	ctx, err := svc.Inspect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ctx.IsGitRepository {
		t.Error("expected IsGitRepository to be true")
	}
	if ctx.CurrentBranch != "main" {
		t.Errorf("expected branch 'main', got %q", ctx.CurrentBranch)
	}
	if ctx.IsDirty {
		t.Error("expected clean state, got dirty")
	}
}

func TestInspect_NotGitRepository(t *testing.T) {
	dir := t.TempDir()

	svc := service.NewProjectInspectionService()
	_, err := svc.Inspect(dir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestInspect_DirtyState(t *testing.T) {
	dir := setupGitRepo(t)

	// Create an untracked file to make it dirty
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := service.NewProjectInspectionService()
	ctx, err := svc.Inspect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ctx.IsDirty {
		t.Error("expected dirty state")
	}
}

func TestInspect_DetectsGoCommands(t *testing.T) {
	dir := setupGitRepo(t)

	// Create go.mod to trigger Go detection
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := service.NewProjectInspectionService()
	ctx, err := svc.Inspect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.TestCommand != "go test ./..." {
		t.Errorf("expected test command 'go test ./...', got %q", ctx.TestCommand)
	}
	if ctx.LintCommand != "golangci-lint run" {
		t.Errorf("expected lint command 'golangci-lint run', got %q", ctx.LintCommand)
	}
	if ctx.TypeCheckCommand != "go build ./..." {
		t.Errorf("expected typecheck command 'go build ./...', got %q", ctx.TypeCheckCommand)
	}
}

func TestInspect_DetectsNodeCommands(t *testing.T) {
	dir := setupGitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := service.NewProjectInspectionService()
	ctx, err := svc.Inspect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.TestCommand != "npm test" {
		t.Errorf("expected test command 'npm test', got %q", ctx.TestCommand)
	}
}

func TestInspect_NoCommandsDetected(t *testing.T) {
	dir := setupGitRepo(t)

	svc := service.NewProjectInspectionService()
	ctx, err := svc.Inspect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.TestCommand != "" {
		t.Errorf("expected empty test command, got %q", ctx.TestCommand)
	}
	if ctx.LintCommand != "" {
		t.Errorf("expected empty lint command, got %q", ctx.LintCommand)
	}
	if ctx.TypeCheckCommand != "" {
		t.Errorf("expected empty typecheck command, got %q", ctx.TypeCheckCommand)
	}
}
