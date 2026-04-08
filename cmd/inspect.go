package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		return nil
	},
}

func init() {
	inspectCmd.Flags().StringVar(&inspectPath, "path", "", "target directory path (default: current directory)")
	rootCmd.AddCommand(inspectCmd)
}
