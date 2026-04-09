package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

var (
	issueTitle       string
	issueDescription string
	issueCategory    string
	issueRemediation string
	issuePriority    int
	issueSource      string
	issueSourceRef   string
	issueDedupKey    string
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues",
}

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue from external source",
	RunE:  runIssueCreate,
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate inputs
	if err := validateIssueInputs(cfg); err != nil {
		return err
	}

	db, err := repository.OpenDatabase(config.DatabasePath(path))
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	issueRepo := repository.NewImplementationIssueRepository(db)
	memoryRepo := repository.NewImprovementMemoryRepository(db)

	// Dedup check
	if issueDedupKey != "" {
		existing, err := issueRepo.FindByDedupKey(issueDedupKey)
		if err != nil {
			return fmt.Errorf("dedup check failed: %w", err)
		}
		if existing != nil {
			existing.IssueDescription = issueDescription
			existing.IssuePriority = issuePriority
			if err := issueRepo.Save(existing); err != nil {
				return fmt.Errorf("failed to update issue: %w", err)
			}
			return outputJSON(existing.IssueId, "updated")
		}
	}

	issue := &domain.ImplementationIssue{
		IssueId:          ulid.MustNew(ulid.Now(), rand.Reader).String(),
		IssueTitle:       issueTitle,
		IssueDescription: issueDescription,
		IssueCategory:    issueCategory,
		RemediationType:  issueRemediation,
		IssuePriority:    issuePriority,
		IssueStatus:      domain.IssueStatusOpen,
		Source:           issueSource,
		SourceRef:        issueSourceRef,
		DedupKey:         issueDedupKey,
		AttemptCount:     0,
		CreatedAt:        time.Now(),
	}

	if err := issueRepo.Save(issue); err != nil {
		return fmt.Errorf("failed to save issue: %w", err)
	}

	ensureMemoryEntry(memoryRepo, issue.IssueCategory)

	return outputJSON(issue.IssueId, "created")
}

func validateIssueInputs(cfg *config.Config) error {
	if strings.TrimSpace(issueTitle) == "" {
		return fmt.Errorf("--title is required and must not be empty")
	}
	if strings.TrimSpace(issueDescription) == "" {
		return fmt.Errorf("--description is required and must not be empty")
	}

	if issueRemediation != domain.RemediationTypeCodePatch && issueRemediation != domain.RemediationTypeConfigPatch {
		return fmt.Errorf("--remediation must be %q or %q", domain.RemediationTypeCodePatch, domain.RemediationTypeConfigPatch)
	}

	// Validate category against allowlist if configured
	if len(cfg.Issues.AllowedCategories) > 0 {
		allowed := false
		for _, c := range cfg.Issues.AllowedCategories {
			if issueCategory == c {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("category %q is not in allowed categories: %v", issueCategory, cfg.Issues.AllowedCategories)
		}
	}

	// Validate priority range if configured
	if cfg.Issues.MaxPriority > 0 && (issuePriority < 1 || issuePriority > cfg.Issues.MaxPriority) {
		return fmt.Errorf("priority must be between 1 and %d", cfg.Issues.MaxPriority)
	}

	// Validate description length if configured
	if cfg.Issues.MaxDescriptionLength > 0 && len(issueDescription) > cfg.Issues.MaxDescriptionLength {
		return fmt.Errorf("description exceeds max length of %d characters", cfg.Issues.MaxDescriptionLength)
	}

	return nil
}

func outputJSON(issueID, status string) error {
	result := map[string]string{
		"issue_id": issueID,
		"status":   status,
	}
	out, err := json.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func init() {
	issueCreateCmd.Flags().StringVar(&issueTitle, "title", "", "issue title (required)")
	issueCreateCmd.Flags().StringVar(&issueDescription, "description", "", "issue description (required)")
	issueCreateCmd.Flags().StringVar(&issueCategory, "category", "", "issue category (required)")
	issueCreateCmd.Flags().StringVar(&issueRemediation, "remediation", domain.RemediationTypeCodePatch, "remediation type: code_patch or config_patch")
	issueCreateCmd.Flags().IntVar(&issuePriority, "priority", 5, "issue priority (lower = higher)")
	issueCreateCmd.Flags().StringVar(&issueSource, "source", domain.IssueSourceExternal, "issue source identifier")
	issueCreateCmd.Flags().StringVar(&issueSourceRef, "source-ref", "", "external reference (rule id, script name)")
	issueCreateCmd.Flags().StringVar(&issueDedupKey, "dedup-key", "", "dedup key for idempotent creation")

	issueCmd.AddCommand(issueCreateCmd)
	rootCmd.AddCommand(issueCmd)
}
