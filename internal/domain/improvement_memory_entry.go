package domain

import "time"

// ImprovementMemoryEntry tracks historical success/failure patterns.
type ImprovementMemoryEntry struct {
	MemoryId           string    `json:"memory_id"`
	PatternKey         string    `json:"pattern_key"`
	PatternDescription string    `json:"pattern_description"`
	SuccessCount       int       `json:"success_count"`
	FailureCount       int       `json:"failure_count"`
	LastObservedAt     time.Time `json:"last_observed_at"`
}
