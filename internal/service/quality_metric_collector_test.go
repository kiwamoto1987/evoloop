package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestCollect_AllCommandsSucceed(t *testing.T) {
	dir := t.TempDir()

	ctx := &domain.ProjectContext{
		ProjectRootPath:  dir,
		TestCommand:      "true",
		LintCommand:      "true",
		TypeCheckCommand: "true",
	}

	collector := service.NewQualityMetricCollector()
	snapshot := collector.Collect(ctx)

	if !snapshot.TestSucceeded {
		t.Error("expected TestSucceeded to be true")
	}
	if !snapshot.LintSucceeded {
		t.Error("expected LintSucceeded to be true")
	}
	if !snapshot.TypeCheckSucceeded {
		t.Error("expected TypeCheckSucceeded to be true")
	}
}

func TestCollect_TestCommandFails(t *testing.T) {
	dir := t.TempDir()

	// Create a script that fails
	script := filepath.Join(dir, "fail.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'FAIL: TestSomething'\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := &domain.ProjectContext{
		ProjectRootPath:  dir,
		TestCommand:      script,
		LintCommand:      "true",
		TypeCheckCommand: "true",
	}

	collector := service.NewQualityMetricCollector()
	snapshot := collector.Collect(ctx)

	if snapshot.TestSucceeded {
		t.Error("expected TestSucceeded to be false")
	}
	if snapshot.TestOutput == "" {
		t.Error("expected TestOutput to contain output")
	}
}

func TestCollect_NoCommandsConfigured(t *testing.T) {
	dir := t.TempDir()

	ctx := &domain.ProjectContext{
		ProjectRootPath: dir,
	}

	collector := service.NewQualityMetricCollector()
	snapshot := collector.Collect(ctx)

	if !snapshot.TestSucceeded {
		t.Error("expected TestSucceeded to be true when no command configured")
	}
	if !snapshot.LintSucceeded {
		t.Error("expected LintSucceeded to be true when no command configured")
	}
	if !snapshot.TypeCheckSucceeded {
		t.Error("expected TypeCheckSucceeded to be true when no command configured")
	}
}
