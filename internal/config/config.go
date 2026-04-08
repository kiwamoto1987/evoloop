package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"gopkg.in/yaml.v3"
)

// Config represents the .evoloop/config.yaml structure.
type Config struct {
	ProjectName string     `yaml:"project_name"`
	LLM         LLMConfig  `yaml:"llm"`
	Evaluation  EvalConfig `yaml:"evaluation"`
	Policies    Policies   `yaml:"policies"`
}

// LLMConfig holds LLM provider settings.
type LLMConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Command  string `yaml:"command"`
}

// EvalConfig holds evaluation command settings.
type EvalConfig struct {
	TestCommand      string `yaml:"test_command"`
	LintCommand      string `yaml:"lint_command"`
	TypeCheckCommand string `yaml:"typecheck_command"`
}

// Policies holds execution policy settings.
type Policies struct {
	MaxChangedFiles int      `yaml:"max_changed_files"`
	MaxChangedLines int      `yaml:"max_changed_lines"`
	DenyPaths       []string `yaml:"deny_paths"`
}

// Load reads config from the given project root.
func Load(projectRoot string) (*Config, error) {
	path := filepath.Join(projectRoot, ".evoloop", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// ApplyEvalCommands overrides ProjectContext commands with config values when set.
func (c *Config) ApplyEvalCommands(ctx *domain.ProjectContext) {
	if c.Evaluation.TestCommand != "" {
		ctx.TestCommand = c.Evaluation.TestCommand
	}
	if c.Evaluation.LintCommand != "" {
		ctx.LintCommand = c.Evaluation.LintCommand
	}
	if c.Evaluation.TypeCheckCommand != "" {
		ctx.TypeCheckCommand = c.Evaluation.TypeCheckCommand
	}
}

// ToExecutionPolicy converts config policies to an ExecutionPolicy.
func (c *Config) ToExecutionPolicy() *policy.ExecutionPolicy {
	p := policy.DefaultPolicy()
	if c.Policies.MaxChangedFiles > 0 {
		p.MaxFiles = c.Policies.MaxChangedFiles
	}
	if c.Policies.MaxChangedLines > 0 {
		p.MaxLines = c.Policies.MaxChangedLines
	}
	if len(c.Policies.DenyPaths) > 0 {
		p.DenyPaths = c.Policies.DenyPaths
	}
	return p
}

// RuntimePath returns the path to the runtime directory.
func RuntimePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".evoloop", "runtime")
}

// DatabasePath returns the path to the SQLite database.
func DatabasePath(projectRoot string) string {
	return filepath.Join(RuntimePath(projectRoot), "improvement.db")
}
