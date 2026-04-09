package repository

import (
	"database/sql"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// HookExecutionRepository handles persistence for HookExecutionRecord.
type HookExecutionRepository struct {
	db *sql.DB
}

// NewHookExecutionRepository creates a new repository.
func NewHookExecutionRepository(db *sql.DB) *HookExecutionRepository {
	return &HookExecutionRepository{db: db}
}

// Save inserts a hook execution record.
func (r *HookExecutionRepository) Save(record *domain.HookExecutionRecord) error {
	_, err := r.db.Exec(
		`INSERT INTO hook_execution_records
		(hook_id, execution_id, hook_type, command, args, exit_code, stdout, stderr, duration_ms, timed_out, executed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.HookId,
		record.ExecutionId,
		record.HookType,
		record.Command,
		strings.Join(record.Args, ","),
		record.ExitCode,
		record.Stdout,
		record.Stderr,
		record.DurationMs,
		record.TimedOut,
		record.ExecutedAt,
	)
	return err
}

// FindByExecutionID retrieves hook records for a given execution.
func (r *HookExecutionRepository) FindByExecutionID(executionID string) ([]*domain.HookExecutionRecord, error) {
	rows, err := r.db.Query(
		`SELECT hook_id, execution_id, hook_type, command, args, exit_code, stdout, stderr, duration_ms, timed_out, executed_at
		FROM hook_execution_records WHERE execution_id = ? ORDER BY executed_at`,
		executionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.HookExecutionRecord
	for rows.Next() {
		record := &domain.HookExecutionRecord{}
		var args string
		if err := rows.Scan(
			&record.HookId, &record.ExecutionId, &record.HookType,
			&record.Command, &args,
			&record.ExitCode, &record.Stdout, &record.Stderr,
			&record.DurationMs, &record.TimedOut, &record.ExecutedAt,
		); err != nil {
			return nil, err
		}
		if args != "" {
			record.Args = strings.Split(args, ",")
		}
		records = append(records, record)
	}

	return records, rows.Err()
}
