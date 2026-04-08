package cmd

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/kiwamoto1987/evoloop/internal/service"
	"github.com/oklog/ulid/v2"
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

		// Open DB for memory lookup and issue saving
		db, err := repository.OpenDatabase(config.DatabasePath(path))
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		memoryRepo := repository.NewImprovementMemoryRepository(db)
		analyzer := service.NewSelfImprovementAnalysisService(memoryRepo)
		issues := analyzer.Analyze(snapshot)

		if len(issues) == 0 {
			fmt.Println("No issues found. All quality checks passed.")
			return nil
		}

		// Ensure memory entries exist for each issue category
		for _, issue := range issues {
			ensureMemoryEntry(memoryRepo, issue.IssueCategory)
		}

		// Save issues to DB
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

func ensureMemoryEntry(repo *repository.ImprovementMemoryRepository, patternKey string) {
	_, err := repo.FindByPatternKey(patternKey)
	if err != nil {
		// Entry doesn't exist, create it
		entry := &domain.ImprovementMemoryEntry{
			MemoryId:           ulid.MustNew(ulid.Now(), rand.Reader).String(),
			PatternKey:         patternKey,
			PatternDescription: patternKey,
			SuccessCount:       0,
			FailureCount:       0,
			LastObservedAt:     time.Now(),
		}
		_ = repo.Save(entry)
	}
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
