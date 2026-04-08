package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// ProjectInspectionService inspects a local Git project.
type ProjectInspectionService struct{}

// NewProjectInspectionService creates a new ProjectInspectionService.
func NewProjectInspectionService() *ProjectInspectionService {
	return &ProjectInspectionService{}
}

// Inspect analyzes the given directory and returns a ProjectContext.
func (s *ProjectInspectionService) Inspect(path string) (*domain.ProjectContext, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if !isGitRepository(absPath) {
		return nil, fmt.Errorf("%s is not a git repository", absPath)
	}

	ctx := &domain.ProjectContext{
		ProjectRootPath: absPath,
		IsGitRepository: true,
	}

	ctx.CurrentBranch = detectBranch(absPath)
	ctx.IsDirty = detectDirtyState(absPath)
	ctx.TestCommand = detectTestCommand(absPath)
	ctx.LintCommand = detectLintCommand(absPath)
	ctx.TypeCheckCommand = detectTypeCheckCommand(absPath)

	return ctx, nil
}

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func detectBranch(path string) string {
	out, err := runGitCommand(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func detectDirtyState(path string) bool {
	out, err := runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func detectTestCommand(path string) string {
	if fileExists(filepath.Join(path, "go.mod")) {
		return "go test ./..."
	}
	if fileExists(filepath.Join(path, "package.json")) {
		return "npm test"
	}
	if fileExists(filepath.Join(path, "pyproject.toml")) {
		return "pytest"
	}
	return ""
}

func detectLintCommand(path string) string {
	if fileExists(filepath.Join(path, "go.mod")) {
		return "golangci-lint run"
	}
	if fileExists(filepath.Join(path, "package.json")) {
		return "eslint ."
	}
	if fileExists(filepath.Join(path, "pyproject.toml")) {
		return "ruff check ."
	}
	return ""
}

func detectTypeCheckCommand(path string) string {
	if fileExists(filepath.Join(path, "go.mod")) {
		return "go build ./..."
	}
	if fileExists(filepath.Join(path, "tsconfig.json")) {
		return "tsc --noEmit"
	}
	if fileExists(filepath.Join(path, "pyproject.toml")) {
		return "mypy ."
	}
	return ""
}

func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
