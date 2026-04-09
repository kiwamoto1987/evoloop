package repository

import (
	"database/sql"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// EvaluationReportRepository handles persistence for EvaluationReport.
type EvaluationReportRepository struct {
	db *sql.DB
}

// NewEvaluationReportRepository creates a new repository.
func NewEvaluationReportRepository(db *sql.DB) *EvaluationReportRepository {
	return &EvaluationReportRepository{db: db}
}

// Save inserts or replaces an evaluation report.
func (r *EvaluationReportRepository) Save(report *domain.EvaluationReport) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO evaluation_reports
		(evaluation_id, execution_id, evaluation_mode, test_status, lint_status, typecheck_status, validate_status, changed_file_count, changed_line_count, evaluation_decision, failure_reasons, generated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		report.EvaluationId, report.ExecutionId,
		report.EvaluationMode,
		report.TestStatus, report.LintStatus, report.TypeCheckStatus, report.ValidateStatus,
		report.ChangedFileCount, report.ChangedLineCount,
		report.EvaluationDecision,
		strings.Join(report.FailureReasons, "|"),
		report.GeneratedAt,
	)
	return err
}

// FindAll retrieves all evaluation reports.
func (r *EvaluationReportRepository) FindAll() ([]*domain.EvaluationReport, error) {
	rows, err := r.db.Query(
		`SELECT evaluation_id, execution_id, evaluation_mode, test_status, lint_status, typecheck_status, validate_status, changed_file_count, changed_line_count, evaluation_decision, failure_reasons, generated_at
		FROM evaluation_reports ORDER BY generated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var reports []*domain.EvaluationReport
	for rows.Next() {
		report := &domain.EvaluationReport{}
		var failureReasons string
		if err := rows.Scan(
			&report.EvaluationId, &report.ExecutionId,
			&report.EvaluationMode,
			&report.TestStatus, &report.LintStatus, &report.TypeCheckStatus, &report.ValidateStatus,
			&report.ChangedFileCount, &report.ChangedLineCount,
			&report.EvaluationDecision, &failureReasons, &report.GeneratedAt,
		); err != nil {
			return nil, err
		}
		if failureReasons != "" {
			report.FailureReasons = strings.Split(failureReasons, "|")
		}
		reports = append(reports, report)
	}

	return reports, rows.Err()
}
