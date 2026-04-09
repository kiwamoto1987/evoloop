package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
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

func TestApplyEvalCommands_OverridesAll(t *testing.T) {
	cfg := &config.Config{
		Evaluation: config.EvalConfig{
			TestCommand:      "custom-test",
			LintCommand:      "custom-lint",
			TypeCheckCommand: "custom-typecheck",
		},
	}

	ctx := &domain.ProjectContext{
		TestCommand:      "auto-detected-test",
		LintCommand:      "auto-detected-lint",
		TypeCheckCommand: "auto-detected-typecheck",
	}

	cfg.ApplyEvalCommands(ctx)

	if ctx.TestCommand != "custom-test" {
		t.Errorf("expected 'custom-test', got %q", ctx.TestCommand)
	}
	if ctx.LintCommand != "custom-lint" {
		t.Errorf("expected 'custom-lint', got %q", ctx.LintCommand)
	}
	if ctx.TypeCheckCommand != "custom-typecheck" {
		t.Errorf("expected 'custom-typecheck', got %q", ctx.TypeCheckCommand)
	}
}

func TestApplyEvalCommands_EmptyKeepsAutoDetected(t *testing.T) {
	cfg := &config.Config{
		Evaluation: config.EvalConfig{
			TestCommand: "custom-test",
			// LintCommand and TypeCheckCommand are empty
		},
	}

	ctx := &domain.ProjectContext{
		TestCommand:      "auto-detected-test",
		LintCommand:      "auto-detected-lint",
		TypeCheckCommand: "auto-detected-typecheck",
	}

	cfg.ApplyEvalCommands(ctx)

	if ctx.TestCommand != "custom-test" {
		t.Errorf("expected 'custom-test', got %q", ctx.TestCommand)
	}
	if ctx.LintCommand != "auto-detected-lint" {
		t.Errorf("expected 'auto-detected-lint' (preserved), got %q", ctx.LintCommand)
	}
	if ctx.TypeCheckCommand != "auto-detected-typecheck" {
		t.Errorf("expected 'auto-detected-typecheck' (preserved), got %q", ctx.TypeCheckCommand)
	}
}

func TestDatabasePath(t *testing.T) {
	path := config.DatabasePath("/project")
	expected := filepath.Join("/project", ".evoloop", "runtime", "improvement.db")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestLoad_FullConfigWithModeBFields(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".evoloop")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `project_name: dex-bot
llm:
  provider: claude
  model: sonnet
  command: "claude"
evaluation:
  test_command: "go test ./..."
  lint_command: "golangci-lint run"
  typecheck_command: "go build ./..."
  validate_commands:
    - "yamllint config.yaml"
    - "./scripts/validate_config.sh"
policies:
  max_changed_files: 5
  max_changed_lines: 200
  deny_paths:
    - ".github/**"
  evaluation_mode: "validate_only"
  max_attempts: 5
  cooldown_minutes: 30
hooks:
  post_apply:
    command: "systemctl"
    args:
      - "restart"
      - "trade-bot"
    timeout_sec: 30
    allowlist:
      - "systemctl"
issues:
  allowed_categories:
    - "kpi_degradation"
    - "config_tuning"
  max_priority: 10
  max_description_length: 5000
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate commands
	if len(cfg.Evaluation.ValidateCommands) != 2 {
		t.Errorf("expected 2 validate_commands, got %d", len(cfg.Evaluation.ValidateCommands))
	}
	if cfg.Evaluation.ValidateCommands[0] != "yamllint config.yaml" {
		t.Errorf("expected first validate command 'yamllint config.yaml', got %q", cfg.Evaluation.ValidateCommands[0])
	}

	// Policies
	if cfg.Policies.EvaluationMode != "validate_only" {
		t.Errorf("expected evaluation_mode 'validate_only', got %q", cfg.Policies.EvaluationMode)
	}
	if cfg.Policies.MaxAttempts != 5 {
		t.Errorf("expected max_attempts 5, got %d", cfg.Policies.MaxAttempts)
	}
	if cfg.Policies.CooldownMinutes != 30 {
		t.Errorf("expected cooldown_minutes 30, got %d", cfg.Policies.CooldownMinutes)
	}

	// Hooks
	if cfg.Hooks.PostApply.Command != "systemctl" {
		t.Errorf("expected hook command 'systemctl', got %q", cfg.Hooks.PostApply.Command)
	}
	if len(cfg.Hooks.PostApply.Args) != 2 || cfg.Hooks.PostApply.Args[0] != "restart" {
		t.Errorf("unexpected hook args: %v", cfg.Hooks.PostApply.Args)
	}
	if cfg.Hooks.PostApply.TimeoutSec != 30 {
		t.Errorf("expected hook timeout 30, got %d", cfg.Hooks.PostApply.TimeoutSec)
	}
	if len(cfg.Hooks.PostApply.Allowlist) != 1 || cfg.Hooks.PostApply.Allowlist[0] != "systemctl" {
		t.Errorf("unexpected hook allowlist: %v", cfg.Hooks.PostApply.Allowlist)
	}

	// Issues
	if len(cfg.Issues.AllowedCategories) != 2 {
		t.Errorf("expected 2 allowed categories, got %d", len(cfg.Issues.AllowedCategories))
	}
	if cfg.Issues.MaxPriority != 10 {
		t.Errorf("expected max_priority 10, got %d", cfg.Issues.MaxPriority)
	}
	if cfg.Issues.MaxDescriptionLength != 5000 {
		t.Errorf("expected max_description_length 5000, got %d", cfg.Issues.MaxDescriptionLength)
	}
}

func TestToExecutionPolicy_WithModeBFields(t *testing.T) {
	cfg := &config.Config{
		Policies: config.Policies{
			MaxChangedFiles:  3,
			MaxChangedLines:  100,
			EvaluationMode:   "validate_only",
			MaxAttempts:      5,
			CooldownMinutes:  30,
		},
	}

	p := cfg.ToExecutionPolicy()
	if p.EvaluationMode != "validate_only" {
		t.Errorf("expected EvaluationMode 'validate_only', got %q", p.EvaluationMode)
	}
	if p.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", p.MaxAttempts)
	}
	if p.CooldownMinutes != 30 {
		t.Errorf("expected CooldownMinutes 30, got %d", p.CooldownMinutes)
	}
}

func TestToExecutionPolicy_DefaultsForNewFields(t *testing.T) {
	cfg := &config.Config{}
	p := cfg.ToExecutionPolicy()
	if p.EvaluationMode != "sandbox" {
		t.Errorf("expected default EvaluationMode 'sandbox', got %q", p.EvaluationMode)
	}
	if p.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts 3, got %d", p.MaxAttempts)
	}
	if p.CooldownMinutes != 60 {
		t.Errorf("expected default CooldownMinutes 60, got %d", p.CooldownMinutes)
	}
}
