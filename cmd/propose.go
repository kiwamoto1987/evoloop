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

var proposeIssueId string

var proposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Generate a patch proposal for an issue using LLM",
	RunE: func(cmd *cobra.Command, args []string) error {
		if proposeIssueId == "" {
			return fmt.Errorf("--issue flag is required")
		}

		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Open DB
		db, err := repository.OpenDatabase(config.DatabasePath(path))
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Load issue from DB
		issueRepo := repository.NewImplementationIssueRepository(db)
		issue, err := issueRepo.FindByID(proposeIssueId)
		if err != nil {
			return fmt.Errorf("issue not found: %w", err)
		}

		// Run proposal
		artifactsPath := config.RuntimePath(path)
		client := llm.NewClaudeCLIClient("claude")
		svc := service.NewImplementationProposalService(client, artifactsPath)

		record, err := svc.Propose(issue, path)
		if err != nil {
			return err
		}

		// Save execution record to DB
		execRepo := repository.NewExecutionHistoryRepository(db)
		if err := execRepo.Save(record); err != nil {
			return fmt.Errorf("failed to save execution record: %w", err)
		}

		// Update issue status
		if record.ExecutionStatus == domain.ExecutionStatusCompleted {
			issue.IssueStatus = domain.IssueStatusProposed
		}
		if err := issueRepo.Save(issue); err != nil {
			return fmt.Errorf("failed to update issue status: %w", err)
		}

		out, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		fmt.Println(string(out))
		return nil
	},
}

func init() {
	proposeCmd.Flags().StringVar(&proposeIssueId, "issue", "", "issue ID to generate proposal for")
	rootCmd.AddCommand(proposeCmd)
}
