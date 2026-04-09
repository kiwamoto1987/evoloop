package domain_test

import (
	"testing"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

func TestCheckStatusConstants(t *testing.T) {
	statuses := []string{
		domain.CheckStatusPassed,
		domain.CheckStatusFailed,
		domain.CheckStatusSkipped,
	}

	seen := make(map[string]bool)
	for _, s := range statuses {
		if s == "" {
			t.Error("check status constant should not be empty")
		}
		if seen[s] {
			t.Errorf("duplicate check status constant: %q", s)
		}
		seen[s] = true
	}
}

func TestEvaluationReportFields(t *testing.T) {
	report := &domain.EvaluationReport{
		EvaluationId:       "eval-1",
		ExecutionId:        "exec-1",
		EvaluationMode:     "sandbox",
		TestStatus:         domain.CheckStatusPassed,
		LintStatus:         domain.CheckStatusFailed,
		TypeCheckStatus:    domain.CheckStatusSkipped,
		ValidateStatus:     domain.CheckStatusSkipped,
		ChangedFileCount:   3,
		ChangedLineCount:   50,
		EvaluationDecision: domain.EvaluationDecisionRejected,
		FailureReasons:     []string{"lint failed"},
		GeneratedAt:        time.Now(),
	}

	if report.EvaluationMode != "sandbox" {
		t.Errorf("EvaluationMode = %q, want %q", report.EvaluationMode, "sandbox")
	}
	if report.TestStatus != domain.CheckStatusPassed {
		t.Errorf("TestStatus = %q, want %q", report.TestStatus, domain.CheckStatusPassed)
	}
	if report.ValidateStatus != domain.CheckStatusSkipped {
		t.Errorf("ValidateStatus = %q, want %q", report.ValidateStatus, domain.CheckStatusSkipped)
	}
}

func TestEvaluationReportValidateOnlyMode(t *testing.T) {
	report := &domain.EvaluationReport{
		EvaluationMode:  "validate_only",
		TestStatus:      domain.CheckStatusSkipped,
		LintStatus:      domain.CheckStatusSkipped,
		TypeCheckStatus: domain.CheckStatusSkipped,
		ValidateStatus:  domain.CheckStatusPassed,
	}

	if report.TestStatus != domain.CheckStatusSkipped {
		t.Errorf("in validate_only mode, TestStatus should be skipped, got %q", report.TestStatus)
	}
	if report.ValidateStatus != domain.CheckStatusPassed {
		t.Errorf("in validate_only mode, ValidateStatus should be passed, got %q", report.ValidateStatus)
	}
}
