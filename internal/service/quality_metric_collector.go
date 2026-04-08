package service

import (
	"os/exec"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// QualityMetricCollector runs quality check commands and collects results.
type QualityMetricCollector struct{}

// NewQualityMetricCollector creates a new QualityMetricCollector.
func NewQualityMetricCollector() *QualityMetricCollector {
	return &QualityMetricCollector{}
}

// Collect runs the configured quality commands and returns a snapshot.
func (c *QualityMetricCollector) Collect(ctx *domain.ProjectContext) *domain.QualityMetricSnapshot {
	snapshot := &domain.QualityMetricSnapshot{}

	if ctx.TestCommand != "" {
		if !isToolAvailable(ctx.TestCommand) {
			snapshot.TestToolMissing = true
			snapshot.TestSucceeded = false
			snapshot.TestOutput = "tool not found: " + strings.Fields(ctx.TestCommand)[0]
		} else {
			ok, output := runCommand(ctx.ProjectRootPath, ctx.TestCommand)
			snapshot.TestSucceeded = ok
			snapshot.TestOutput = output
		}
	} else {
		snapshot.TestSucceeded = true
	}

	if ctx.LintCommand != "" {
		if !isToolAvailable(ctx.LintCommand) {
			snapshot.LintToolMissing = true
			snapshot.LintSucceeded = false
			snapshot.LintOutput = "tool not found: " + strings.Fields(ctx.LintCommand)[0]
		} else {
			ok, output := runCommand(ctx.ProjectRootPath, ctx.LintCommand)
			snapshot.LintSucceeded = ok
			snapshot.LintOutput = output
		}
	} else {
		snapshot.LintSucceeded = true
	}

	if ctx.TypeCheckCommand != "" {
		if !isToolAvailable(ctx.TypeCheckCommand) {
			snapshot.TypeCheckToolMissing = true
			snapshot.TypeCheckSucceeded = false
			snapshot.TypeCheckOutput = "tool not found: " + strings.Fields(ctx.TypeCheckCommand)[0]
		} else {
			ok, output := runCommand(ctx.ProjectRootPath, ctx.TypeCheckCommand)
			snapshot.TypeCheckSucceeded = ok
			snapshot.TypeCheckOutput = output
		}
	} else {
		snapshot.TypeCheckSucceeded = true
	}

	return snapshot
}

func isToolAvailable(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	_, err := exec.LookPath(parts[0])
	return err == nil
}

func runCommand(dir, command string) (bool, string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return true, ""
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}
