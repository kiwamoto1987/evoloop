package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"github.com/kiwamoto1987/evoloop/internal/repository"
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

		// Open DB
		db, err := repository.OpenDatabase(config.DatabasePath(path))
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Load execution record from DB
		execRepo := repository.NewExecutionHistoryRepository(db)
		record, err := execRepo.FindByID(evaluateExecutionId)
		if err != nil {
			return fmt.Errorf("execution record not found: %w", err)
		}

		// Run evaluation
		svc := service.NewSelfImprovementEvaluationService(policy.DefaultPolicy())
		report, err := svc.Evaluate(record, projectCtx)
		if err != nil {
			return err
		}

		// Save evaluation report to DB
		evalRepo := repository.NewEvaluationReportRepository(db)
		if err := evalRepo.Save(report); err != nil {
			return fmt.Errorf("failed to save evaluation report: %w", err)
		}

		// Update issue status based on evaluation decision
		issueRepo := repository.NewImplementationIssueRepository(db)
		issue, err := issueRepo.FindByID(record.IssueId)
		if err == nil {
			if report.EvaluationDecision == domain.EvaluationDecisionAccepted {
				issue.IssueStatus = domain.IssueStatusAccepted
			} else {
				issue.IssueStatus = domain.IssueStatusRejected
			}
			if err := issueRepo.Save(issue); err != nil {
				return fmt.Errorf("failed to update issue status: %w", err)
			}
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
