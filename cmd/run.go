package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/llm"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/spf13/cobra"
)

var (
	runMaxIterations int
	runMaxFailures   int
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full self-improvement loop automatically",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Load config
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Open DB
		db, err := repository.OpenDatabase(config.DatabasePath(path))
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		issueRepo := repository.NewImplementationIssueRepository(db)
		execRepo := repository.NewExecutionHistoryRepository(db)
		evalRepo := repository.NewEvaluationReportRepository(db)
		memoryRepo := repository.NewImprovementMemoryRepository(db)

		consecutiveFailures := 0

		for iteration := 1; iteration <= runMaxIterations; iteration++ {
			fmt.Printf("\n=== Iteration %d/%d ===\n", iteration, runMaxIterations)

			// Step 1: Inspect
			inspectionSvc := service.NewProjectInspectionService()
			projectCtx, err := inspectionSvc.Inspect(path)
			if err != nil {
				return fmt.Errorf("inspect failed: %w", err)
			}
			cfg.ApplyEvalCommands(projectCtx)

			// Step 2: Analyze
			fmt.Println("[analyze] Running quality checks...")
			collector := service.NewQualityMetricCollector()
			snapshot := collector.Collect(projectCtx)

			analyzer := service.NewSelfImprovementAnalysisService(memoryRepo)
			issues := analyzer.Analyze(snapshot)

			// Save issues
			for _, issue := range issues {
				ensureMemoryEntry(memoryRepo, issue.IssueCategory)
				if err := issueRepo.Save(issue); err != nil {
					return fmt.Errorf("failed to save issue: %w", err)
				}
			}

			// Select highest priority proposable issue
			selected := selectBestIssue(issues)
			if selected == nil {
				fmt.Println("[analyze] No proposable issues found. Loop complete.")
				break
			}
			fmt.Printf("[analyze] Selected issue: %s (%s)\n", selected.IssueTitle, selected.IssueCategory)

			// Step 3: Propose
			fmt.Println("[propose] Generating patch...")
			client := llm.NewClaudeCLIClient(cfg.LLM.Command)
			proposalSvc := service.NewImplementationProposalService(client, config.RuntimePath(path))

			record, err := proposalSvc.Propose(selected, path)
			if err != nil {
				fmt.Printf("[propose] Failed: %v\n", err)
				if record != nil {
					_ = execRepo.Save(record)
				}
				consecutiveFailures++
				if consecutiveFailures >= runMaxFailures {
					fmt.Printf("[stop] %d consecutive failures. Stopping.\n", consecutiveFailures)
					break
				}
				continue
			}

			if err := execRepo.Save(record); err != nil {
				return fmt.Errorf("failed to save execution record: %w", err)
			}

			if record.ExecutionStatus != domain.ExecutionStatusCompleted {
				fmt.Printf("[propose] Status: %s\n", record.ExecutionStatus)
				consecutiveFailures++
				if consecutiveFailures >= runMaxFailures {
					fmt.Printf("[stop] %d consecutive failures. Stopping.\n", consecutiveFailures)
					break
				}
				continue
			}

			selected.IssueStatus = domain.IssueStatusProposed
			_ = issueRepo.Save(selected)

			// Step 4: Evaluate
			fmt.Println("[evaluate] Evaluating patch...")
			evalSvc := service.NewSelfImprovementEvaluationService(cfg.ToExecutionPolicy())
			report, err := evalSvc.Evaluate(record, projectCtx)
			if err != nil {
				fmt.Printf("[evaluate] Error: %v\n", err)
				consecutiveFailures++
				if consecutiveFailures >= runMaxFailures {
					fmt.Printf("[stop] %d consecutive failures. Stopping.\n", consecutiveFailures)
					break
				}
				continue
			}

			if err := evalRepo.Save(report); err != nil {
				return fmt.Errorf("failed to save evaluation: %w", err)
			}

			if report.EvaluationDecision == domain.EvaluationDecisionAccepted {
				selected.IssueStatus = domain.IssueStatusAccepted
				_ = memoryRepo.RecordSuccess(selected.IssueCategory)
				consecutiveFailures = 0
				fmt.Println("[evaluate] ✓ Accepted")
			} else {
				selected.IssueStatus = domain.IssueStatusRejected
				_ = memoryRepo.RecordFailure(selected.IssueCategory)
				consecutiveFailures++
				fmt.Printf("[evaluate] ✗ Rejected: %v\n", report.FailureReasons)
			}
			_ = issueRepo.Save(selected)

			if consecutiveFailures >= runMaxFailures {
				fmt.Printf("[stop] %d consecutive failures. Stopping.\n", consecutiveFailures)
				break
			}

			// Print summary
			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))
		}

		fmt.Println("\n=== Run complete ===")
		return nil
	},
}

func selectBestIssue(issues []*domain.ImplementationIssue) *domain.ImplementationIssue {
	var best *domain.ImplementationIssue
	for _, issue := range issues {
		if !issue.IsProposable() {
			continue
		}
		if best == nil || issue.IssuePriority < best.IssuePriority {
			best = issue
		}
	}
	return best
}

func init() {
	runCmd.Flags().IntVar(&runMaxIterations, "max-iterations", 1, "maximum number of improvement iterations")
	runCmd.Flags().IntVar(&runMaxFailures, "max-failures", 3, "stop after this many consecutive failures")
	rootCmd.AddCommand(runCmd)
}
