package llm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// ClaudeCLIClient implements LanguageModelClient using the Claude CLI.
type ClaudeCLIClient struct {
	Command string
}

// NewClaudeCLIClient creates a new ClaudeCLIClient.
func NewClaudeCLIClient(command string) *ClaudeCLIClient {
	if command == "" {
		command = "claude"
	}
	return &ClaudeCLIClient{Command: command}
}

// GeneratePatch calls the Claude CLI to generate a patch for the given context.
func (c *ClaudeCLIClient) GeneratePatch(input *domain.PromptContext) (*domain.PatchResult, error) {
	prompt := buildPrompt(input)

	cmd := exec.Command(c.Command, "-p", prompt)
	cmd.Dir = input.ProjectRootPath
	out, err := cmd.CombinedOutput()
	rawOutput := string(out)

	if err != nil {
		return nil, fmt.Errorf("claude CLI failed: %w\noutput: %s", err, rawOutput)
	}

	return &domain.PatchResult{
		PatchContent: extractPatch(rawOutput),
		RawOutput:    rawOutput,
	}, nil
}

func buildPrompt(input *domain.PromptContext) string {
	var b strings.Builder

	b.WriteString("Fix the following issue in this project.\n\n")
	b.WriteString(fmt.Sprintf("Issue: %s\n", input.IssueTitle))
	b.WriteString(fmt.Sprintf("Description: %s\n", input.IssueDescription))

	if len(input.AcceptanceCriteria) > 0 {
		b.WriteString("\nAcceptance Criteria:\n")
		for _, ac := range input.AcceptanceCriteria {
			b.WriteString(fmt.Sprintf("- %s\n", ac))
		}
	}

	if len(input.TargetPaths) > 0 {
		b.WriteString(fmt.Sprintf("\nTarget paths: %s\n", strings.Join(input.TargetPaths, ", ")))
	}

	b.WriteString("\nConstraints:\n")
	b.WriteString("- Keep changes minimal\n")
	b.WriteString("- Do not add new dependencies\n")
	b.WriteString("- Do not delete tests\n")
	b.WriteString("- Output ONLY a unified diff patch, no explanation\n")

	return b.String()
}

func extractPatch(output string) string {
	// Look for unified diff markers
	lines := strings.Split(output, "\n")
	var patchLines []string
	inPatch := false

	for _, line := range lines {
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "--- ") {
			inPatch = true
		}
		if inPatch {
			patchLines = append(patchLines, line)
		}
	}

	if len(patchLines) > 0 {
		return strings.Join(patchLines, "\n")
	}

	// If no diff markers found, return the full output as patch content
	return strings.TrimSpace(output)
}
