package policy

import "path/filepath"

// ExecutionPolicy defines constraints for patch evaluation.
type ExecutionPolicy struct {
	MaxFiles        int
	MaxLines        int
	DenyPaths       []string
	EvaluationMode  string // "sandbox" | "validate_only"
	MaxAttempts     int
	CooldownMinutes int
}

// DefaultPolicy returns the default execution policy.
func DefaultPolicy() *ExecutionPolicy {
	return &ExecutionPolicy{
		MaxFiles: 5,
		MaxLines: 200,
		DenyPaths: []string{
			".github/**",
			".evoloop/**",
		},
		EvaluationMode:  "sandbox",
		MaxAttempts:     3,
		CooldownMinutes: 60,
	}
}

// CheckFileCount returns true if the file count is within limits.
func (p *ExecutionPolicy) CheckFileCount(count int) bool {
	return count <= p.MaxFiles
}

// CheckLineCount returns true if the line count is within limits.
func (p *ExecutionPolicy) CheckLineCount(count int) bool {
	return count <= p.MaxLines
}

// CheckDenyPaths returns true if none of the changed paths match deny patterns.
func (p *ExecutionPolicy) CheckDenyPaths(changedPaths []string) bool {
	for _, changed := range changedPaths {
		for _, pattern := range p.DenyPaths {
			if matched, _ := filepath.Match(pattern, changed); matched {
				return false
			}
		}
	}
	return true
}
