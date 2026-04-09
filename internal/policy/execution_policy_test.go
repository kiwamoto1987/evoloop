package policy_test

import (
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/policy"
)

func TestDefaultPolicy(t *testing.T) {
	p := policy.DefaultPolicy()

	if p.MaxFiles != 5 {
		t.Errorf("expected MaxFiles 5, got %d", p.MaxFiles)
	}
	if p.MaxLines != 200 {
		t.Errorf("expected MaxLines 200, got %d", p.MaxLines)
	}
}

func TestCheckFileCount_WithinLimit(t *testing.T) {
	p := policy.DefaultPolicy()
	if !p.CheckFileCount(3) {
		t.Error("expected 3 files to be within limit")
	}
}

func TestCheckFileCount_ExceedsLimit(t *testing.T) {
	p := policy.DefaultPolicy()
	if p.CheckFileCount(6) {
		t.Error("expected 6 files to exceed limit")
	}
}

func TestCheckLineCount_WithinLimit(t *testing.T) {
	p := policy.DefaultPolicy()
	if !p.CheckLineCount(100) {
		t.Error("expected 100 lines to be within limit")
	}
}

func TestCheckLineCount_ExceedsLimit(t *testing.T) {
	p := policy.DefaultPolicy()
	if p.CheckLineCount(201) {
		t.Error("expected 201 lines to exceed limit")
	}
}

func TestCheckFileCount_AtLimit(t *testing.T) {
	p := policy.DefaultPolicy()
	if !p.CheckFileCount(5) {
		t.Error("expected 5 files (at limit) to pass")
	}
}

func TestDefaultPolicy_NewFields(t *testing.T) {
	p := policy.DefaultPolicy()

	if p.EvaluationMode != "sandbox" {
		t.Errorf("expected EvaluationMode 'sandbox', got %q", p.EvaluationMode)
	}
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", p.MaxAttempts)
	}
	if p.CooldownMinutes != 60 {
		t.Errorf("expected CooldownMinutes 60, got %d", p.CooldownMinutes)
	}
}

func TestCheckDenyPaths_NoDeny(t *testing.T) {
	p := policy.DefaultPolicy()
	if !p.CheckDenyPaths([]string{"src/main.go", "internal/service/foo.go"}) {
		t.Error("expected non-denied paths to pass")
	}
}

func TestCheckDenyPaths_Denied(t *testing.T) {
	p := &policy.ExecutionPolicy{
		DenyPaths: []string{".github/**", ".evoloop/**"},
	}

	// Direct match with glob pattern — filepath.Match only matches single-level
	if p.CheckDenyPaths([]string{".github/workflows"}) {
		t.Error("expected .github/workflows to be denied")
	}
}

func TestCheckDenyPaths_Empty(t *testing.T) {
	p := &policy.ExecutionPolicy{
		DenyPaths: []string{},
	}
	if !p.CheckDenyPaths([]string{"anything.go"}) {
		t.Error("expected empty deny list to pass everything")
	}
}

func TestCheckDenyPaths_NoChangedPaths(t *testing.T) {
	p := policy.DefaultPolicy()
	if !p.CheckDenyPaths([]string{}) {
		t.Error("expected no changed paths to pass")
	}
}
