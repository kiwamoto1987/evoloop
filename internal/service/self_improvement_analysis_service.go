package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/oklog/ulid/v2"
)

// SelfImprovementAnalysisService generates ImplementationIssues from quality metrics.
type SelfImprovementAnalysisService struct{}

// NewSelfImprovementAnalysisService creates a new SelfImprovementAnalysisService.
func NewSelfImprovementAnalysisService() *SelfImprovementAnalysisService {
	return &SelfImprovementAnalysisService{}
}

// Analyze generates issues from a quality metric snapshot.
func (s *SelfImprovementAnalysisService) Analyze(snapshot *domain.QualityMetricSnapshot) []*domain.ImplementationIssue {
	var issues []*domain.ImplementationIssue

	if !snapshot.TestSucceeded {
		issues = append(issues, &domain.ImplementationIssue{
			IssueId:          generateID(),
			IssueTitle:       "Fix test failures",
			IssueDescription: fmt.Sprintf("Test command failed.\n\nOutput:\n%s", truncateOutput(snapshot.TestOutput)),
			IssueCategory:    domain.IssueCategoryTestFailure,
			IssuePriority:    1,
			IssueStatus:      domain.IssueStatusOpen,
			AcceptanceCriteria: []string{
				"All tests pass",
				"No new test failures introduced",
			},
			CreatedAt: time.Now(),
		})
	}

	if !snapshot.LintSucceeded {
		issues = append(issues, &domain.ImplementationIssue{
			IssueId:          generateID(),
			IssueTitle:       "Fix lint violations",
			IssueDescription: fmt.Sprintf("Lint command failed.\n\nOutput:\n%s", truncateOutput(snapshot.LintOutput)),
			IssueCategory:    domain.IssueCategoryLintViolation,
			IssuePriority:    2,
			IssueStatus:      domain.IssueStatusOpen,
			AcceptanceCriteria: []string{
				"All lint checks pass",
				"No new lint violations introduced",
			},
			CreatedAt: time.Now(),
		})
	}

	if !snapshot.TypeCheckSucceeded {
		issues = append(issues, &domain.ImplementationIssue{
			IssueId:          generateID(),
			IssueTitle:       "Fix type check errors",
			IssueDescription: fmt.Sprintf("Type check command failed.\n\nOutput:\n%s", truncateOutput(snapshot.TypeCheckOutput)),
			IssueCategory:    domain.IssueCategoryTypeCheckFailure,
			IssuePriority:    1,
			IssueStatus:      domain.IssueStatusOpen,
			AcceptanceCriteria: []string{
				"Type check passes",
				"No new type errors introduced",
			},
			CreatedAt: time.Now(),
		})
	}

	return issues
}

func generateID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}

func truncateOutput(output string) string {
	const maxLen = 2000
	if len(output) > maxLen {
		return output[:maxLen] + "\n... (truncated)"
	}
	return output
}
