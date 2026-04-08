package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/llm"
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

		// TODO: load issue from DB once persistence is implemented
		issue := &domain.ImplementationIssue{
			IssueId:    proposeIssueId,
			IssueTitle: "Issue " + proposeIssueId,
		}

		artifactsPath := filepath.Join(path, ".evoloop", "runtime")
		client := llm.NewClaudeCLIClient("claude")
		svc := service.NewImplementationProposalService(client, artifactsPath)

		record, err := svc.Propose(issue, path)
		if err != nil {
			return err
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
