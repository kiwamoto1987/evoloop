package service

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
	"github.com/kiwamoto1987/evoloop/internal/policy"
	"github.com/oklog/ulid/v2"
)

// CopyExcludePatterns defines directories to skip when copying a project.
var CopyExcludePatterns = []string{
	".git",
	"node_modules",
	"vendor",
	".evoloop",
}

// SelfImprovementEvaluationService evaluates a patch proposal.
type SelfImprovementEvaluationService struct {
	policy *policy.ExecutionPolicy
}

// NewSelfImprovementEvaluationService creates a new evaluation service.
func NewSelfImprovementEvaluationService(p *policy.ExecutionPolicy) *SelfImprovementEvaluationService {
	return &SelfImprovementEvaluationService{policy: p}
}

// Evaluate runs the evaluation pipeline for a given execution record and project context.
// validateCommands is used only in validate_only mode; pass nil for sandbox mode.
func (s *SelfImprovementEvaluationService) Evaluate(
	record *domain.ExecutionRecord,
	projectCtx *domain.ProjectContext,
	validateCommands []string,
) (*domain.EvaluationReport, error) {
	report := &domain.EvaluationReport{
		EvaluationId: ulid.MustNew(ulid.Now(), rand.Reader).String(),
		ExecutionId:  record.ExecutionId,
		GeneratedAt:  time.Now(),
	}

	// Read patch content
	patchContent, err := os.ReadFile(record.PatchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "evoloop-eval-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Copy project to temp directory
	if err := copyProject(projectCtx.ProjectRootPath, tmpDir); err != nil {
		return nil, fmt.Errorf("failed to copy project: %w", err)
	}

	// Apply patch
	if err := applyPatch(tmpDir, string(patchContent)); err != nil {
		report.EvaluationDecision = domain.EvaluationDecisionRejected
		report.FailureReasons = append(report.FailureReasons, fmt.Sprintf("patch apply failed: %v", err))
		return report, nil
	}

	// Count changes
	fileCount, lineCount := countChanges(string(patchContent))
	report.ChangedFileCount = fileCount
	report.ChangedLineCount = lineCount

	var failures []string

	switch s.policy.EvaluationMode {
	case "validate_only":
		failures = s.evaluateValidateOnly(tmpDir, validateCommands, report)
	default:
		failures = s.evaluateSandbox(tmpDir, projectCtx, report)
	}

	// Check policy constraints
	if !s.policy.CheckFileCount(fileCount) {
		failures = append(failures, fmt.Sprintf("changed files %d exceeds limit %d", fileCount, s.policy.MaxFiles))
	}
	if !s.policy.CheckLineCount(lineCount) {
		failures = append(failures, fmt.Sprintf("changed lines %d exceeds limit %d", lineCount, s.policy.MaxLines))
	}

	// Decide
	if len(failures) == 0 {
		report.EvaluationDecision = domain.EvaluationDecisionAccepted
	} else {
		report.EvaluationDecision = domain.EvaluationDecisionRejected
		report.FailureReasons = failures
	}

	return report, nil
}

func (s *SelfImprovementEvaluationService) evaluateSandbox(
	tmpDir string,
	projectCtx *domain.ProjectContext,
	report *domain.EvaluationReport,
) []string {
	report.EvaluationMode = "sandbox"
	var failures []string

	if projectCtx.TestCommand != "" && isToolInstalled(projectCtx.TestCommand) {
		ok, _ := runCommandInDir(tmpDir, projectCtx.TestCommand)
		if ok {
			report.TestStatus = domain.CheckStatusPassed
		} else {
			report.TestStatus = domain.CheckStatusFailed
			failures = append(failures, "tests failed")
		}
	} else {
		report.TestStatus = domain.CheckStatusSkipped
	}

	if projectCtx.LintCommand != "" && isToolInstalled(projectCtx.LintCommand) {
		ok, _ := runCommandInDir(tmpDir, projectCtx.LintCommand)
		if ok {
			report.LintStatus = domain.CheckStatusPassed
		} else {
			report.LintStatus = domain.CheckStatusFailed
			failures = append(failures, "lint failed")
		}
	} else {
		report.LintStatus = domain.CheckStatusSkipped
	}

	if projectCtx.TypeCheckCommand != "" && isToolInstalled(projectCtx.TypeCheckCommand) {
		ok, _ := runCommandInDir(tmpDir, projectCtx.TypeCheckCommand)
		if ok {
			report.TypeCheckStatus = domain.CheckStatusPassed
		} else {
			report.TypeCheckStatus = domain.CheckStatusFailed
			failures = append(failures, "typecheck failed")
		}
	} else {
		report.TypeCheckStatus = domain.CheckStatusSkipped
	}

	report.ValidateStatus = domain.CheckStatusSkipped
	return failures
}

func (s *SelfImprovementEvaluationService) evaluateValidateOnly(
	tmpDir string,
	validateCommands []string,
	report *domain.EvaluationReport,
) []string {
	report.EvaluationMode = "validate_only"
	report.TestStatus = domain.CheckStatusSkipped
	report.LintStatus = domain.CheckStatusSkipped
	report.TypeCheckStatus = domain.CheckStatusSkipped

	var failures []string

	if len(validateCommands) == 0 {
		report.ValidateStatus = domain.CheckStatusSkipped
		return failures
	}

	allPassed := true
	for _, cmd := range validateCommands {
		if cmd == "" {
			continue
		}
		if !isToolInstalled(cmd) {
			continue
		}
		ok, _ := runCommandInDir(tmpDir, cmd)
		if !ok {
			allPassed = false
			failures = append(failures, fmt.Sprintf("validate command failed: %s", cmd))
		}
	}

	if allPassed {
		report.ValidateStatus = domain.CheckStatusPassed
	} else {
		report.ValidateStatus = domain.CheckStatusFailed
	}

	return failures
}

func copyProject(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check exclusions
		for _, pattern := range CopyExcludePatterns {
			if strings.HasPrefix(relPath, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func applyPatch(dir, patchContent string) error {
	patchFile := filepath.Join(dir, ".evoloop-patch.tmp")
	if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
		return err
	}
	defer func() { _ = os.Remove(patchFile) }()

	cmd := exec.Command("patch", "-p1", "-i", patchFile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

func countChanges(patchContent string) (files, lines int) {
	fileSet := make(map[string]bool)
	for _, line := range strings.Split(patchContent, "\n") {
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "--- a/") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fileSet[parts[len(parts)-1]] = true
			}
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			lines++
		}
	}
	return len(fileSet), lines
}

func isToolInstalled(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	_, err := exec.LookPath(parts[0])
	return err == nil
}

func runCommandInDir(dir, command string) (bool, string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return true, ""
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}
