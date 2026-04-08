package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/spf13/cobra"
)

var inspectPath string

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect project structure and detect available commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := inspectPath
		if path == "" {
			var err error
			path, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		svc := service.NewProjectInspectionService()
		ctx, err := svc.Inspect(path)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		fmt.Println(string(out))

		// Save artifact
		reportsDir := filepath.Join(config.RuntimePath(path), "reports")
		if err := os.MkdirAll(reportsDir, 0755); err != nil {
			return fmt.Errorf("failed to create reports directory: %w", err)
		}
		artifactPath := filepath.Join(reportsDir, "project_inspection.json")
		if err := os.WriteFile(artifactPath, out, 0644); err != nil {
			return fmt.Errorf("failed to save inspection report: %w", err)
		}

		return nil
	},
}

func init() {
	inspectCmd.Flags().StringVar(&inspectPath, "path", "", "target directory path (default: current directory)")
	rootCmd.AddCommand(inspectCmd)
}
