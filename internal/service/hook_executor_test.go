package service_test

import (
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestHookExecutor_Success(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "echo",
		Args:       []string{"hello"},
		TimeoutSec: 5,
		Allowlist:  []string{"echo"},
	}

	record, err := executor.Execute(hook, "EXEC001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", record.ExitCode)
	}
	if record.TimedOut {
		t.Error("expected no timeout")
	}
	if record.Command != "echo" {
		t.Errorf("expected command 'echo', got %q", record.Command)
	}
	if record.HookType != "post_apply" {
		t.Errorf("expected hook type 'post_apply', got %q", record.HookType)
	}
	if record.ExecutionId != "EXEC001" {
		t.Errorf("expected execution ID 'EXEC001', got %q", record.ExecutionId)
	}
	if record.DurationMs < 0 {
		t.Errorf("expected non-negative duration, got %d", record.DurationMs)
	}
	if record.HookId == "" {
		t.Error("expected non-empty HookId")
	}
}

func TestHookExecutor_CommandFailure(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "false",
		Args:       []string{},
		TimeoutSec: 5,
		Allowlist:  []string{"false"},
	}

	record, err := executor.Execute(hook, "EXEC002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
	if record.TimedOut {
		t.Error("expected no timeout for failed command")
	}
}

func TestHookExecutor_AllowlistRejection(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "rm",
		Args:       []string{"-rf", "/"},
		TimeoutSec: 5,
		Allowlist:  []string{"systemctl", "echo"},
	}

	_, err := executor.Execute(hook, "EXEC003")
	if err == nil {
		t.Fatal("expected error for command not in allowlist")
	}
}

func TestHookExecutor_Timeout(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "sleep",
		Args:       []string{"10"},
		TimeoutSec: 1,
		Allowlist:  []string{"sleep"},
	}

	record, err := executor.Execute(hook, "EXEC004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !record.TimedOut {
		t.Error("expected timeout")
	}
}

func TestHookExecutor_CapturesStdout(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "echo",
		Args:       []string{"test output"},
		TimeoutSec: 5,
		Allowlist:  []string{"echo"},
	}

	record, err := executor.Execute(hook, "EXEC005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.Stdout != "test output\n" {
		t.Errorf("expected stdout 'test output\\n', got %q", record.Stdout)
	}
}

func TestHookExecutor_DefaultTimeout(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "echo",
		Args:       []string{"hi"},
		TimeoutSec: 0, // should use default
		Allowlist:  []string{"echo"},
	}

	record, err := executor.Execute(hook, "EXEC006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", record.ExitCode)
	}
}

func TestHookExecutor_EmptyAllowlist(t *testing.T) {
	executor := service.NewHookExecutor()
	hook := config.PostApplyHook{
		Command:    "echo",
		Args:       []string{"hi"},
		TimeoutSec: 5,
		Allowlist:  []string{},
	}

	_, err := executor.Execute(hook, "EXEC007")
	if err == nil {
		t.Fatal("expected error when allowlist is empty")
	}
}
