package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
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

		// Load config for policy settings
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Run evaluation
		svc := service.NewSelfImprovementEvaluationService(cfg.ToExecutionPolicy())
		report, err := svc.Evaluate(record, projectCtx)
		if err != nil {
			return err
		}

		// Save evaluation report to DB
		evalRepo := repository.NewEvaluationReportRepository(db)
		if err := evalRepo.Save(report); err != nil {
			return fmt.Errorf("failed to save evaluation report: %w", err)
		}

		// Update issue status and improvement memory
		issueRepo := repository.NewImplementationIssueRepository(db)
		memoryRepo := repository.NewImprovementMemoryRepository(db)
		issue, err := issueRepo.FindByID(record.IssueId)
		if err == nil {
			if report.EvaluationDecision == domain.EvaluationDecisionAccepted {
				issue.IssueStatus = domain.IssueStatusAccepted
				_ = memoryRepo.RecordSuccess(issue.IssueCategory)
			} else {
				issue.IssueStatus = domain.IssueStatusRejected
				_ = memoryRepo.RecordFailure(issue.IssueCategory)
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
