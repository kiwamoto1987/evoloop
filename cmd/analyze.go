package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze project quality and generate improvement issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		inspectionSvc := service.NewProjectInspectionService()
		ctx, err := inspectionSvc.Inspect(path)
		if err != nil {
			return fmt.Errorf("inspect failed: %w", err)
		}

		collector := service.NewQualityMetricCollector()
		snapshot := collector.Collect(ctx)

		analyzer := service.NewSelfImprovementAnalysisService()
		issues := analyzer.Analyze(snapshot)

		if len(issues) == 0 {
			fmt.Println("No issues found. All quality checks passed.")
			return nil
		}

		// Save issues to DB
		db, err := repository.OpenDatabase(config.DatabasePath(path))
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		issueRepo := repository.NewImplementationIssueRepository(db)
		for _, issue := range issues {
			if err := issueRepo.Save(issue); err != nil {
				return fmt.Errorf("failed to save issue: %w", err)
			}
		}

		out, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal issues: %w", err)
		}

		fmt.Println(string(out))
		fmt.Printf("\n%d issue(s) saved to database.\n", len(issues))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
