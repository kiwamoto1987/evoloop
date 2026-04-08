package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".evoloop")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `project_name: evoloop
llm:
  provider: claude
  model: sonnet
  command: "claude"
evaluation:
  test_command: "go test ./..."
  lint_command: "golangci-lint run"
  typecheck_command: "go build ./..."
policies:
  max_changed_files: 5
  max_changed_lines: 200
  deny_paths:
    - ".github/**"
    - ".evoloop/**"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectName != "evoloop" {
		t.Errorf("expected project_name 'evoloop', got %q", cfg.ProjectName)
	}
	if cfg.LLM.Provider != "claude" {
		t.Errorf("expected provider 'claude', got %q", cfg.LLM.Provider)
	}
	if cfg.Evaluation.TestCommand != "go test ./..." {
		t.Errorf("expected test_command 'go test ./...', got %q", cfg.Evaluation.TestCommand)
	}
	if cfg.Policies.MaxChangedFiles != 5 {
		t.Errorf("expected max_changed_files 5, got %d", cfg.Policies.MaxChangedFiles)
	}
	if len(cfg.Policies.DenyPaths) != 2 {
		t.Errorf("expected 2 deny paths, got %d", len(cfg.Policies.DenyPaths))
	}
}

func TestLoad_MissingConfig(t *testing.T) {
	dir := t.TempDir()
	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestToExecutionPolicy(t *testing.T) {
	cfg := &config.Config{
		Policies: config.Policies{
			MaxChangedFiles: 10,
			MaxChangedLines: 300,
			DenyPaths:       []string{".secret/**"},
		},
	}

	p := cfg.ToExecutionPolicy()
	if p.MaxFiles != 10 {
		t.Errorf("expected MaxFiles 10, got %d", p.MaxFiles)
	}
	if p.MaxLines != 300 {
		t.Errorf("expected MaxLines 300, got %d", p.MaxLines)
	}
	if len(p.DenyPaths) != 1 || p.DenyPaths[0] != ".secret/**" {
		t.Errorf("unexpected DenyPaths: %v", p.DenyPaths)
	}
}

func TestToExecutionPolicy_Defaults(t *testing.T) {
	cfg := &config.Config{}
	p := cfg.ToExecutionPolicy()
	if p.MaxFiles != 5 {
		t.Errorf("expected default MaxFiles 5, got %d", p.MaxFiles)
	}
	if p.MaxLines != 200 {
		t.Errorf("expected default MaxLines 200, got %d", p.MaxLines)
	}
}

func TestDatabasePath(t *testing.T) {
	path := config.DatabasePath("/project")
	expected := filepath.Join("/project", ".evoloop", "runtime", "improvement.db")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
