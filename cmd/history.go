package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/spf13/cobra"
)

var historyJSON bool

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show execution history",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		dbPath := filepath.Join(path, ".evoloop", "runtime", "improvement.db")
		db, err := repository.OpenDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		issueRepo := repository.NewImplementationIssueRepository(db)
		executionRepo := repository.NewExecutionHistoryRepository(db)
		evaluationRepo := repository.NewEvaluationReportRepository(db)

		svc := service.NewExecutionHistoryQueryService(issueRepo, executionRepo, evaluationRepo)
		summary, err := svc.QueryAll()
		if err != nil {
			return err
		}

		if historyJSON {
			out, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal history: %w", err)
			}
			fmt.Println(string(out))
			return nil
		}

		printTable(summary)
		return nil
	},
}

func printTable(summary *service.HistorySummary) {
	if len(summary.Issues) == 0 && len(summary.Executions) == 0 && len(summary.Evaluations) == 0 {
		fmt.Println("No history found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if len(summary.Issues) > 0 {
		fmt.Fprintln(w, "ISSUES")
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tSTATUS\tPRIORITY")
		for _, issue := range summary.Issues {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
				issue.IssueId, issue.IssueTitle, issue.IssueCategory, issue.IssueStatus, issue.IssuePriority)
		}
		fmt.Fprintln(w)
	}

	if len(summary.Executions) > 0 {
		fmt.Fprintln(w, "EXECUTIONS")
		fmt.Fprintln(w, "ID\tISSUE\tSTATUS\tMODEL\tSTARTED")
		for _, exec := range summary.Executions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s/%s\t%s\n",
				exec.ExecutionId, exec.IssueId, exec.ExecutionStatus,
				exec.ModelProvider, exec.ModelName,
				exec.StartedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Fprintln(w)
	}

	if len(summary.Evaluations) > 0 {
		fmt.Fprintln(w, "EVALUATIONS")
		fmt.Fprintln(w, "ID\tEXECUTION\tDECISION\tFILES\tLINES")
		for _, eval := range summary.Evaluations {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
				eval.EvaluationId, eval.ExecutionId, eval.EvaluationDecision,
				eval.ChangedFileCount, eval.ChangedLineCount)
		}
	}

	w.Flush()
}

func init() {
	historyCmd.Flags().BoolVar(&historyJSON, "json", false, "output in JSON format")
	rootCmd.AddCommand(historyCmd)
}
