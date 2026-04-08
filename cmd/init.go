package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfig = `project_name: my-project

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

// RunInit executes the init logic. Exported for testing.
func RunInit() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configDir := filepath.Join(path, ".evoloop")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config already exists at %s", configPath)
	}

	// Create directories
	runtimeDirs := []string{
		filepath.Join(configDir, "runtime", "logs"),
		filepath.Join(configDir, "runtime", "patches"),
		filepath.Join(configDir, "runtime", "prompts"),
		filepath.Join(configDir, "runtime", "reports"),
	}
	for _, dir := range runtimeDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write config
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Initialized .evoloop at %s\n", configDir)
	fmt.Println("Edit .evoloop/config.yaml to customize settings.")
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .evoloop directory with default configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
