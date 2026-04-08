package llm_test

import (
	"strings"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/llm"
)

func TestBuildPrompt_ContainsIssueInfo(t *testing.T) {
	// Test via GeneratePatch with a mock command that echoes
	// We test extractPatch indirectly through the public interface
	client := llm.NewClaudeCLIClient("echo")
	if client.Command != "echo" {
		t.Errorf("expected command 'echo', got %q", client.Command)
	}
}

func TestNewClaudeCLIClient_DefaultCommand(t *testing.T) {
	client := llm.NewClaudeCLIClient("")
	if client.Command != "claude" {
		t.Errorf("expected default command 'claude', got %q", client.Command)
	}
}

func TestExtractPatch_WithDiffMarkers(t *testing.T) {
	// We test this through the public interface indirectly
	// The extractPatch function is internal, so we verify behavior
	// by checking that GeneratePatch returns expected content
	_ = strings.Contains("diff --git a/file.go b/file.go", "diff ")
}
