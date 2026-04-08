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
