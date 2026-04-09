package domain

import "time"

// HookExecutionRecord tracks the result of a post-apply hook execution.
type HookExecutionRecord struct {
	HookId      string    `json:"hook_id"`
	ExecutionId string    `json:"execution_id"`
	HookType    string    `json:"hook_type"`
	Command     string    `json:"command"`
	Args        []string  `json:"args,omitempty"`
	ExitCode    int       `json:"exit_code"`
	Stdout      string    `json:"stdout,omitempty"`
	Stderr      string    `json:"stderr,omitempty"`
	DurationMs  int       `json:"duration_ms"`
	TimedOut    bool      `json:"timed_out"`
	ExecutedAt  time.Time `json:"executed_at"`
}
