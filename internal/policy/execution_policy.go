package policy

// ExecutionPolicy defines constraints for patch evaluation.
type ExecutionPolicy struct {
	MaxFiles  int
	MaxLines  int
	DenyPaths []string
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
