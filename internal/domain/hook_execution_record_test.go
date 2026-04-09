package domain_test

import (
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

func TestHookExecutionRecordFields(t *testing.T) {
	record := &domain.HookExecutionRecord{
		HookId:      "hook-1",
		ExecutionId: "exec-1",
		HookType:    "post_apply",
		Command:     "systemctl",
		Args:        []string{"restart", "trade-bot"},
		ExitCode:    0,
		Stdout:      "ok",
		Stderr:      "",
		DurationMs:  1500,
		TimedOut:    false,
		ExecutedAt:  time.Now(),
	}

	if record.HookType != "post_apply" {
		t.Errorf("HookType = %q, want %q", record.HookType, "post_apply")
	}
	if record.Command != "systemctl" {
		t.Errorf("Command = %q, want %q", record.Command, "systemctl")
	}
	if len(record.Args) != 2 {
		t.Errorf("Args length = %d, want 2", len(record.Args))
	}
	if record.TimedOut {
		t.Error("expected TimedOut to be false")
	}
}
