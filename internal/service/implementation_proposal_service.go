package service

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/llm"
	"github.com/oklog/ulid/v2"
)

// ImplementationProposalService generates patch proposals for issues.
type ImplementationProposalService struct {
	client        llm.LanguageModelClient
	artifactsPath string
}

// NewImplementationProposalService creates a new ImplementationProposalService.
func NewImplementationProposalService(client llm.LanguageModelClient, artifactsPath string) *ImplementationProposalService {
	return &ImplementationProposalService{
		client:        client,
		artifactsPath: artifactsPath,
	}
}

// Propose generates a patch proposal for the given issue.
func (s *ImplementationProposalService) Propose(issue *domain.ImplementationIssue, projectRoot string) (*domain.ExecutionRecord, error) {
	executionId := ulid.MustNew(ulid.Now(), rand.Reader).String()
	now := time.Now()

	record := &domain.ExecutionRecord{
		ExecutionId:     executionId,
		IssueId:         issue.IssueId,
		ExecutionStatus: domain.ExecutionStatusPending,
		ModelProvider:   "claude",
		ModelName:       "sonnet",
		StartedAt:       now,
	}

	promptCtx := &domain.PromptContext{
		ProjectRootPath:    projectRoot,
		IssueId:            issue.IssueId,
		IssueTitle:         issue.IssueTitle,
		IssueDescription:   issue.IssueDescription,
		AcceptanceCriteria: issue.AcceptanceCriteria,
		TargetPaths:        issue.TargetPaths,
	}

	// Save prompt artifact
	promptPath, err := s.saveArtifact("prompts", executionId+".txt", buildPromptText(promptCtx))
	if err != nil {
		record.ExecutionStatus = domain.ExecutionStatusFailed
		record.FinishedAt = time.Now()
		return record, fmt.Errorf("failed to save prompt: %w", err)
	}
	record.PromptPath = promptPath

	// Call LLM
	result, err := s.client.GeneratePatch(promptCtx)
	if err != nil {
		record.ExecutionStatus = domain.ExecutionStatusFailed
		record.FinishedAt = time.Now()
		return record, nil
	}

	// Save patch artifact
	patchPath, err := s.saveArtifact("patches", executionId+".patch", result.PatchContent)
	if err != nil {
		record.ExecutionStatus = domain.ExecutionStatusFailed
		record.FinishedAt = time.Now()
		return record, fmt.Errorf("failed to save patch: %w", err)
	}
	record.PatchPath = patchPath

	record.ExecutionStatus = domain.ExecutionStatusCompleted
	record.FinishedAt = time.Now()

	return record, nil
}

func (s *ImplementationProposalService) saveArtifact(subdir, filename, content string) (string, error) {
	dir := filepath.Join(s.artifactsPath, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func buildPromptText(ctx *domain.PromptContext) string {
	return fmt.Sprintf("Issue: %s\nDescription: %s\n", ctx.IssueTitle, ctx.IssueDescription)
}
