package llm

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// Default tools: read-only access so Claude can inspect files but not modify them.
var defaultAllowedTools = []string{"Read", "Glob", "Grep"}

// ClaudeCLIClient implements LanguageModelClient using the Claude CLI.
type ClaudeCLIClient struct {
	Command string
	// AllowedTools specifies tools Claude is permitted to use.
	// Defaults to read-only tools (Read, Glob, Grep).
	// Set to include "Edit","Write","Bash" for future tool-enabled invocations.
	AllowedTools []string
}

// NewClaudeCLIClient creates a new ClaudeCLIClient.
func NewClaudeCLIClient(command string) *ClaudeCLIClient {
	if command == "" {
		command = "claude"
	}
	return &ClaudeCLIClient{
		Command:      command,
		AllowedTools: defaultAllowedTools,
	}
}

// GeneratePatch calls the Claude CLI to generate a patch for the given context.
func (c *ClaudeCLIClient) GeneratePatch(input *domain.PromptContext) (*domain.PatchResult, error) {
	prompt := buildPrompt(input)

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--max-turns", "5",
	}

	for _, tool := range c.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}

	cmd := exec.Command(c.Command, args...)
	cmd.Dir = input.ProjectRootPath
	out, err := cmd.CombinedOutput()
	rawOutput := string(out)

	if err != nil {
		return nil, fmt.Errorf("claude CLI failed: %w\noutput: %s", err, rawOutput)
	}

	// Parse JSON output from Claude CLI
	content := extractContentFromJSON(rawOutput)

	return &domain.PatchResult{
		PatchContent: extractPatch(content),
		RawOutput:    rawOutput,
	}, nil
}

// claudeJSONResponse represents the JSON output from claude --output-format json.
type claudeJSONResponse struct {
	Result string `json:"result"`
}

func extractContentFromJSON(output string) string {
	var resp claudeJSONResponse
	if err := json.Unmarshal([]byte(output), &resp); err == nil && resp.Result != "" {
		return resp.Result
	}
	// Fallback: treat output as plain text
	return output
}

func buildPrompt(input *domain.PromptContext) string {
	var b strings.Builder

	b.WriteString("Fix the following issue in this project.\n\n")
	fmt.Fprintf(&b, "Issue: %s\n", input.IssueTitle)
	fmt.Fprintf(&b, "Description: %s\n", input.IssueDescription)

	if len(input.AcceptanceCriteria) > 0 {
		b.WriteString("\nAcceptance Criteria:\n")
		for _, ac := range input.AcceptanceCriteria {
			fmt.Fprintf(&b, "- %s\n", ac)
		}
	}

	if len(input.TargetPaths) > 0 {
		fmt.Fprintf(&b, "\nTarget paths: %s\n", strings.Join(input.TargetPaths, ", "))
	}

	if len(input.RelevantFileContents) > 0 {
		b.WriteString("\nRelevant files:\n")
		for path, content := range input.RelevantFileContents {
			fmt.Fprintf(&b, "\n--- %s ---\n%s\n", path, content)
		}
	}

	b.WriteString("\nConstraints:\n")
	b.WriteString("- Keep changes minimal\n")
	b.WriteString("- Do not add new dependencies\n")
	b.WriteString("- Do not delete tests\n")
	b.WriteString("- Do NOT modify any files directly\n")
	b.WriteString("- Output ONLY a unified diff patch, no explanation, no markdown code fences\n")

	return b.String()
}

func extractPatch(output string) string {
	lines := strings.Split(output, "\n")
	var patchLines []string
	inPatch := false

	for _, line := range lines {
		// Skip markdown code fences
		if strings.HasPrefix(line, "```") {
			if inPatch {
				break
			}
			continue
		}

		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "--- ") {
			inPatch = true
		}
		if inPatch {
			patchLines = append(patchLines, line)
		}
	}

	if len(patchLines) > 0 {
		// Ensure patch ends with newline
		result := strings.Join(patchLines, "\n")
		if !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		return result
	}

	// If no diff markers found, return the full output as patch content
	return strings.TrimSpace(output)
}
