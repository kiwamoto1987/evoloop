package service

import (
	"sort"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
)

// IssueSelector picks the next issue to propose, respecting retry limits and cooldown.
type IssueSelector struct {
	policy *policy.ExecutionPolicy
}

// NewIssueSelector creates a new IssueSelector.
func NewIssueSelector(p *policy.ExecutionPolicy) *IssueSelector {
	return &IssueSelector{policy: p}
}

// SelectNext returns the best candidate issue, or nil if none is eligible.
// Selection: filter non-proposable, over-max-attempts, and cooldown-blocked issues,
// then sort by priority ASC, attempt count ASC.
func (s *IssueSelector) SelectNext(issues []*domain.ImplementationIssue) *domain.ImplementationIssue {
	now := time.Now()
	var candidates []*domain.ImplementationIssue

	for _, issue := range issues {
		if !issue.IsProposable() {
			continue
		}
		if issue.AttemptCount >= s.policy.MaxAttempts {
			continue
		}
		if issue.AttemptCount > 0 && s.policy.CooldownMinutes > 0 {
			cooldownEnd := issue.LastAttemptedAt.Add(time.Duration(s.policy.CooldownMinutes) * time.Minute)
			if now.Before(cooldownEnd) {
				continue
			}
		}
		candidates = append(candidates, issue)
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].IssuePriority != candidates[j].IssuePriority {
			return candidates[i].IssuePriority < candidates[j].IssuePriority
		}
		return candidates[i].AttemptCount < candidates[j].AttemptCount
	})

	return candidates[0]
}
