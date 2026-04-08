package service_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

// mockLLMClient is a test double for LanguageModelClient.
type mockLLMClient struct {
	result *domain.PatchResult
	err    error
}

func (m *mockLLMClient) GeneratePatch(input *domain.PromptContext) (*domain.PatchResult, error) {
	return m.result, m.err
}

func TestPropose_Success(t *testing.T) {
	dir := t.TempDir()
	artifactsPath := filepath.Join(dir, "runtime")

	client := &mockLLMClient{
		result: &domain.PatchResult{
			PatchContent: "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n",
			RawOutput:    "some raw output",
		},
	}

	svc := service.NewImplementationProposalService(client, artifactsPath)

	issue := &domain.ImplementationIssue{
		IssueId:            "TEST001",
		IssueTitle:         "Fix test failure",
		IssueDescription:   "Tests are failing",
		AcceptanceCriteria: []string{"Tests pass"},
		CreatedAt:          time.Now(),
	}

	record, err := svc.Propose(issue, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.ExecutionStatus != domain.ExecutionStatusCompleted {
		t.Errorf("expected status %q, got %q", domain.ExecutionStatusCompleted, record.ExecutionStatus)
	}
	if record.IssueId != "TEST001" {
		t.Errorf("expected IssueId 'TEST001', got %q", record.IssueId)
	}
	if record.ExecutionId == "" {
		t.Error("expected ExecutionId to be set")
	}
	if record.PromptPath == "" {
		t.Error("expected PromptPath to be set")
	}
	if record.PatchPath == "" {
		t.Error("expected PatchPath to be set")
	}

	// Verify artifacts exist on disk
	if _, err := os.Stat(record.PromptPath); os.IsNotExist(err) {
		t.Errorf("prompt file does not exist: %s", record.PromptPath)
	}
	if _, err := os.Stat(record.PatchPath); os.IsNotExist(err) {
		t.Errorf("patch file does not exist: %s", record.PatchPath)
	}

	// Verify patch content
	patchContent, err := os.ReadFile(record.PatchPath)
	if err != nil {
		t.Fatalf("failed to read patch: %v", err)
	}
	if string(patchContent) != "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n" {
		t.Errorf("unexpected patch content: %s", patchContent)
	}
}

func TestPropose_LLMFailure(t *testing.T) {
	dir := t.TempDir()
	artifactsPath := filepath.Join(dir, "runtime")

	client := &mockLLMClient{
		result: nil,
		err:    fmt.Errorf("LLM unavailable"),
	}

	svc := service.NewImplementationProposalService(client, artifactsPath)

	issue := &domain.ImplementationIssue{
		IssueId:    "TEST002",
		IssueTitle: "Fix something",
		CreatedAt:  time.Now(),
	}

	record, err := svc.Propose(issue, dir)
	if err != nil {
		t.Fatalf("unexpected error (should be nil, failure recorded in record): %v", err)
	}

	if record.ExecutionStatus != domain.ExecutionStatusFailed {
		t.Errorf("expected status %q, got %q", domain.ExecutionStatusFailed, record.ExecutionStatus)
	}
	if record.PatchPath != "" {
		t.Errorf("expected empty PatchPath on failure, got %q", record.PatchPath)
	}
}

func TestPropose_RecordTimestamps(t *testing.T) {
	dir := t.TempDir()
	artifactsPath := filepath.Join(dir, "runtime")

	client := &mockLLMClient{
		result: &domain.PatchResult{
			PatchContent: "patch",
			RawOutput:    "output",
		},
	}

	svc := service.NewImplementationProposalService(client, artifactsPath)

	issue := &domain.ImplementationIssue{
		IssueId:    "TEST003",
		IssueTitle: "Test timestamps",
		CreatedAt:  time.Now(),
	}

	before := time.Now()
	record, err := svc.Propose(issue, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now()

	if record.StartedAt.Before(before) || record.StartedAt.After(after) {
		t.Errorf("StartedAt %v not in expected range [%v, %v]", record.StartedAt, before, after)
	}
	if record.FinishedAt.Before(record.StartedAt) {
		t.Error("FinishedAt should not be before StartedAt")
	}
}
