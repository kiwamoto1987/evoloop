package llm

import (
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

func TestNewClaudeCLIClient_DefaultCommand(t *testing.T) {
	client := NewClaudeCLIClient("")
	if client.Command != "claude" {
		t.Errorf("expected default command 'claude', got %q", client.Command)
	}
}

func TestNewClaudeCLIClient_CustomCommand(t *testing.T) {
	client := NewClaudeCLIClient("my-claude")
	if client.Command != "my-claude" {
		t.Errorf("expected command 'my-claude', got %q", client.Command)
	}
}

func TestExtractContentFromJSON_ValidJSON(t *testing.T) {
	input := `{"result":"diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go"}`
	content := extractContentFromJSON(input)
	if content != "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestExtractContentFromJSON_InvalidJSON(t *testing.T) {
	input := "plain text output"
	content := extractContentFromJSON(input)
	if content != "plain text output" {
		t.Errorf("expected fallback to plain text, got %q", content)
	}
}

func TestExtractPatch_WithDiffMarkers(t *testing.T) {
	input := "Some explanation\ndiff --git a/file.go b/file.go\n--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@\n-old\n+new"
	patch := extractPatch(input)
	if !contains(patch, "diff --git") {
		t.Errorf("expected patch to contain diff markers, got %q", patch)
	}
	if contains(patch, "Some explanation") {
		t.Error("expected patch to exclude non-diff content")
	}
}

func TestExtractPatch_NoDiffMarkers(t *testing.T) {
	input := "just some text output"
	patch := extractPatch(input)
	if patch != "just some text output" {
		t.Errorf("expected full output as fallback, got %q", patch)
	}
}

func TestBuildPrompt_ContainsIssueInfo(t *testing.T) {
	input := &domain.PromptContext{
		IssueTitle:         "Fix test failure",
		IssueDescription:   "Tests are failing",
		AcceptanceCriteria: []string{"Tests pass"},
		TargetPaths:        []string{"internal/service"},
		RelevantFileContents: map[string]string{
			"main.go": "package main",
		},
	}

	prompt := buildPrompt(input)

	for _, want := range []string{
		"Fix test failure",
		"Tests are failing",
		"Tests pass",
		"internal/service",
		"main.go",
		"package main",
	} {
		if !contains(prompt, want) {
			t.Errorf("expected prompt to contain %q", want)
		}
	}
}

func TestAllowedTools_Empty(t *testing.T) {
	client := NewClaudeCLIClient("claude")
	if len(client.AllowedTools) != 0 {
		t.Errorf("expected empty AllowedTools by default, got %v", client.AllowedTools)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
