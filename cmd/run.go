package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	runAutoApply     bool
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
		defer func() { _ = db.Close() }()

		issueRepo := repository.NewImplementationIssueRepository(db)
		execRepo := repository.NewExecutionHistoryRepository(db)
		evalRepo := repository.NewEvaluationReportRepository(db)
		memoryRepo := repository.NewImprovementMemoryRepository(db)
		hookRepo := repository.NewHookExecutionRepository(db)

		execPolicy := cfg.ToExecutionPolicy()
		selector := service.NewIssueSelector(execPolicy)
		hookExecutor := service.NewHookExecutor()

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

			// Step 2: Select issue — DB-first, then analyze
			var selected *domain.ImplementationIssue

			openIssues, err := issueRepo.FindOpenProposable()
			if err != nil {
				return fmt.Errorf("failed to query open issues: %w", err)
			}

			if len(openIssues) > 0 {
				fmt.Printf("[select] Found %d open proposable issue(s) in DB\n", len(openIssues))
				selected = selector.SelectNext(openIssues)
				// Close issues that exceeded max attempts
				for _, issue := range openIssues {
					if issue.AttemptCount >= execPolicy.MaxAttempts {
						issue.IssueStatus = domain.IssueStatusClosed
						_ = issueRepo.Save(issue)
						fmt.Printf("[select] Closed issue %s (max attempts exceeded)\n", issue.IssueId)
					}
				}
			}

			if selected == nil {
				// No eligible DB issues, run analyze
				fmt.Println("[analyze] Running quality checks...")
				collector := service.NewQualityMetricCollector()
				snapshot := collector.Collect(projectCtx)

				analyzer := service.NewSelfImprovementAnalysisService(memoryRepo)
				issues := analyzer.Analyze(snapshot)

				for _, issue := range issues {
					issue.Source = domain.IssueSourceAnalyze
					issue.RemediationType = domain.RemediationTypeCodePatch
					ensureMemoryEntry(memoryRepo, issue.IssueCategory)
					if err := issueRepo.Save(issue); err != nil {
						return fmt.Errorf("failed to save issue: %w", err)
					}
				}

				selected = selector.SelectNext(issues)
			}

			if selected == nil {
				fmt.Println("[select] No eligible issues found. Loop complete.")
				break
			}
			fmt.Printf("[select] Selected issue: %s (%s)\n", selected.IssueTitle, selected.IssueCategory)

			// Update attempt tracking
			selected.AttemptCount++
			selected.LastAttemptedAt = time.Now()
			selected.IssueStatus = domain.IssueStatusProposed
			_ = issueRepo.Save(selected)

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
				selected.IssueStatus = domain.IssueStatusOpen
				_ = issueRepo.Save(selected)
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
				selected.IssueStatus = domain.IssueStatusOpen
				_ = issueRepo.Save(selected)
				consecutiveFailures++
				if consecutiveFailures >= runMaxFailures {
					fmt.Printf("[stop] %d consecutive failures. Stopping.\n", consecutiveFailures)
					break
				}
				continue
			}

			// Step 4: Evaluate
			fmt.Println("[evaluate] Evaluating patch...")
			evalSvc := service.NewSelfImprovementEvaluationService(execPolicy)
			report, err := evalSvc.Evaluate(record, projectCtx, cfg.Evaluation.ValidateCommands)
			if err != nil {
				fmt.Printf("[evaluate] Error: %v\n", err)
				selected.IssueStatus = domain.IssueStatusOpen
				_ = issueRepo.Save(selected)
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
				_ = issueRepo.Save(selected)
				fmt.Println("[evaluate] Accepted")

				if runAutoApply {
					// Apply patch
					if err := applyPatchToProject(path, record.PatchPath); err != nil {
						fmt.Printf("[apply] Failed: %v\n", err)
						selected.IssueStatus = domain.IssueStatusApplyFailed
						_ = issueRepo.Save(selected)
						_ = memoryRepo.RecordFailure(selected.IssueCategory)
						consecutiveFailures++
					} else {
						selected.IssueStatus = domain.IssueStatusApplied
						_ = issueRepo.Save(selected)
						fmt.Println("[apply] Patch applied to project")

						// Execute post-apply hook
						if cfg.Hooks.PostApply.Command != "" {
							fmt.Printf("[hook] Executing: %s %s\n", cfg.Hooks.PostApply.Command, strings.Join(cfg.Hooks.PostApply.Args, " "))
							hookRecord, hookErr := hookExecutor.Execute(cfg.Hooks.PostApply, record.ExecutionId)
							if hookErr != nil {
								fmt.Printf("[hook] Error: %v\n", hookErr)
								selected.IssueStatus = domain.IssueStatusHookFailed
								_ = issueRepo.Save(selected)
								_ = memoryRepo.RecordFailure(selected.IssueCategory)
								consecutiveFailures++
							} else {
								_ = hookRepo.Save(hookRecord)
								if hookRecord.ExitCode == 0 && !hookRecord.TimedOut {
									selected.IssueStatus = domain.IssueStatusCompleted
									_ = issueRepo.Save(selected)
									_ = memoryRepo.RecordSuccess(selected.IssueCategory)
									consecutiveFailures = 0
									fmt.Println("[hook] Success")
								} else {
									fmt.Printf("[hook] Failed: exit=%d timeout=%v\n", hookRecord.ExitCode, hookRecord.TimedOut)
									selected.IssueStatus = domain.IssueStatusHookFailed
									_ = issueRepo.Save(selected)
									_ = memoryRepo.RecordFailure(selected.IssueCategory)
									consecutiveFailures++
								}
							}
						} else {
							// No hook configured → complete
							selected.IssueStatus = domain.IssueStatusCompleted
							_ = issueRepo.Save(selected)
							_ = memoryRepo.RecordSuccess(selected.IssueCategory)
							consecutiveFailures = 0
						}
					}
				} else {
					// Not auto-apply → just mark accepted
					_ = memoryRepo.RecordSuccess(selected.IssueCategory)
					consecutiveFailures = 0
				}
			} else {
				selected.IssueStatus = domain.IssueStatusRejected
				_ = issueRepo.Save(selected)
				_ = memoryRepo.RecordFailure(selected.IssueCategory)
				consecutiveFailures++
				fmt.Printf("[evaluate] Rejected: %v\n", report.FailureReasons)
			}

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

func applyPatchToProject(projectDir, patchPath string) error {
	patchData, err := os.ReadFile(patchPath)
	if err != nil {
		return fmt.Errorf("failed to read patch: %w", err)
	}

	cmd := exec.Command("patch", "-p1")
	cmd.Dir = projectDir
	cmd.Stdin = strings.NewReader(string(patchData))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("patch command failed: %w\noutput: %s", err, string(out))
	}
	return nil
}

func init() {
	runCmd.Flags().IntVar(&runMaxIterations, "max-iterations", 1, "maximum number of improvement iterations")
	runCmd.Flags().IntVar(&runMaxFailures, "max-failures", 3, "stop after this many consecutive failures")
	runCmd.Flags().BoolVar(&runAutoApply, "auto-apply", false, "automatically apply accepted patches to the project")
	rootCmd.AddCommand(runCmd)
}
