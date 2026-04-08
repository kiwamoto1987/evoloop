package service

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/repository"
	"github.com/oklog/ulid/v2"
)

// SelfImprovementAnalysisService generates ImplementationIssues from quality metrics.
type SelfImprovementAnalysisService struct {
	memoryRepo *repository.ImprovementMemoryRepository
}

// NewSelfImprovementAnalysisService creates a new SelfImprovementAnalysisService.
// memoryRepo can be nil if memory is not available.
func NewSelfImprovementAnalysisService(memoryRepo *repository.ImprovementMemoryRepository) *SelfImprovementAnalysisService {
	return &SelfImprovementAnalysisService{memoryRepo: memoryRepo}
}

// Analyze generates issues from a quality metric snapshot.
func (s *SelfImprovementAnalysisService) Analyze(snapshot *domain.QualityMetricSnapshot) []*domain.ImplementationIssue {
	var issues []*domain.ImplementationIssue

	if snapshot.TestToolMissing {
		issues = append(issues, s.newEnvironmentIssue("Test tool not found", snapshot.TestOutput))
	} else if !snapshot.TestSucceeded {
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

	if snapshot.LintToolMissing {
		issues = append(issues, s.newEnvironmentIssue("Lint tool not found", snapshot.LintOutput))
	} else if !snapshot.LintSucceeded {
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

	if snapshot.TypeCheckToolMissing {
		issues = append(issues, s.newEnvironmentIssue("TypeCheck tool not found", snapshot.TypeCheckOutput))
	} else if !snapshot.TypeCheckSucceeded {
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

	// Apply priority correction from memory
	s.adjustPriorities(issues)

	return issues
}

// adjustPriorities modifies issue priorities based on historical success/failure rates.
// Issues with high failure rates are deprioritized (higher number = lower priority).
// Issues with high success rates are boosted (lower number = higher priority).
func (s *SelfImprovementAnalysisService) adjustPriorities(issues []*domain.ImplementationIssue) {
	if s.memoryRepo == nil {
		return
	}

	for _, issue := range issues {
		entry, err := s.memoryRepo.FindByPatternKey(issue.IssueCategory)
		if err != nil {
			continue
		}

		total := entry.SuccessCount + entry.FailureCount
		if total == 0 {
			continue
		}

		failureRate := float64(entry.FailureCount) / float64(total)
		if failureRate > 0.7 {
			// High failure rate: deprioritize
			issue.IssuePriority += 2
		} else if failureRate < 0.3 {
			// High success rate: boost priority
			if issue.IssuePriority > 1 {
				issue.IssuePriority--
			}
		}
	}
}

func (s *SelfImprovementAnalysisService) newEnvironmentIssue(title, output string) *domain.ImplementationIssue {
	return &domain.ImplementationIssue{
		IssueId:          generateID(),
		IssueTitle:       title,
		IssueDescription: output,
		IssueCategory:    domain.IssueCategoryEnvironment,
		IssuePriority:    0,
		IssueStatus:      domain.IssueStatusOpen,
		AcceptanceCriteria: []string{
			"Install the required tool",
		},
		CreatedAt: time.Now(),
	}
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
