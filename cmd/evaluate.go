package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/spf13/cobra"
)

var evaluateExecutionId string

var evaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate a patch proposal against quality checks and policy constraints",
	RunE: func(cmd *cobra.Command, args []string) error {
		if evaluateExecutionId == "" {
			return fmt.Errorf("--execution flag is required")
		}

		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Inspect current project
		inspectionSvc := service.NewProjectInspectionService()
		projectCtx, err := inspectionSvc.Inspect(path)
		if err != nil {
			return fmt.Errorf("inspect failed: %w", err)
		}

		// TODO: load execution record from DB once persistence is implemented
		record := &domain.ExecutionRecord{
			ExecutionId: evaluateExecutionId,
		}

		svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
		report, err := svc.Evaluate(record, projectCtx)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		fmt.Println(string(out))
		return nil
	},
}

func init() {
	evaluateCmd.Flags().StringVar(&evaluateExecutionId, "execution", "", "execution ID to evaluate")
	rootCmd.AddCommand(evaluateCmd)
}
