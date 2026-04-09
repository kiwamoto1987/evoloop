package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os/exec"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/config"
	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/oklog/ulid/v2"
)

const defaultHookTimeoutSec = 30

// HookExecutor runs post-apply hooks with safety constraints.
type HookExecutor struct{}

// NewHookExecutor creates a new HookExecutor.
func NewHookExecutor() *HookExecutor {
	return &HookExecutor{}
}

// Execute runs the hook command, enforcing allowlist and timeout.
// Returns a HookExecutionRecord with captured output, or an error if the command is not allowed.
func (h *HookExecutor) Execute(hook config.PostApplyHook, executionID string) (*domain.HookExecutionRecord, error) {
	if err := validateAllowlist(hook.Command, hook.Allowlist); err != nil {
		return nil, err
	}

	timeoutSec := hook.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = defaultHookTimeoutSec
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, hook.Command, hook.Args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	record := &domain.HookExecutionRecord{
		HookId:      ulid.MustNew(ulid.Now(), rand.Reader).String(),
		ExecutionId: executionID,
		HookType:    "post_apply",
		Command:     hook.Command,
		Args:        hook.Args,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		DurationMs:  int(duration.Milliseconds()),
		ExecutedAt:  start,
	}

	if ctx.Err() == context.DeadlineExceeded {
		record.TimedOut = true
		record.ExitCode = -1
	} else if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			record.ExitCode = exitErr.ExitCode()
		} else {
			record.ExitCode = -1
		}
	} else {
		record.ExitCode = 0
	}

	return record, nil
}

func validateAllowlist(command string, allowlist []string) error {
	if len(allowlist) == 0 {
		return fmt.Errorf("hook allowlist is empty: no commands are permitted")
	}
	for _, allowed := range allowlist {
		if command == allowed {
			return nil
		}
	}
	return fmt.Errorf("command %q is not in allowlist %v", command, allowlist)
}
