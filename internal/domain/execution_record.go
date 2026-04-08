package domain

import "time"

// ExecutionStatus represents the lifecycle state of an ExecutionRecord.
const (
	ExecutionStatusPending   = "Pending"
	ExecutionStatusCompleted = "Completed"
	ExecutionStatusFailed    = "Failed"
)

// ExecutionRecord tracks a single proposal execution against an issue.
type ExecutionRecord struct {
	ExecutionId string `json:"execution_id"`
	IssueId     string `json:"issue_id"`

	ExecutionStatus string `json:"execution_status"`

	ModelProvider string `json:"model_provider"`
	ModelName     string `json:"model_name"`

	PromptPath string `json:"prompt_path"`
	PatchPath  string `json:"patch_path"`

	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}
