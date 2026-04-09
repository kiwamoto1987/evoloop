package repository

import (
	"database/sql"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// ExecutionHistoryRepository handles persistence for ExecutionRecord.
type ExecutionHistoryRepository struct {
	db *sql.DB
}

// NewExecutionHistoryRepository creates a new repository.
func NewExecutionHistoryRepository(db *sql.DB) *ExecutionHistoryRepository {
	return &ExecutionHistoryRepository{db: db}
}

// Save inserts or replaces an execution record.
func (r *ExecutionHistoryRepository) Save(record *domain.ExecutionRecord) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO execution_records
		(execution_id, issue_id, execution_status, model_provider, model_name, prompt_path, patch_path, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ExecutionId, record.IssueId, record.ExecutionStatus,
		record.ModelProvider, record.ModelName,
		record.PromptPath, record.PatchPath,
		record.StartedAt, record.FinishedAt,
	)
	return err
}

// FindByID retrieves an execution record by ID.
func (r *ExecutionHistoryRepository) FindByID(id string) (*domain.ExecutionRecord, error) {
	row := r.db.QueryRow(
		`SELECT execution_id, issue_id, execution_status, model_provider, model_name, prompt_path, patch_path, started_at, finished_at
		FROM execution_records WHERE execution_id = ?`, id,
	)

	record := &domain.ExecutionRecord{}
	err := row.Scan(
		&record.ExecutionId, &record.IssueId, &record.ExecutionStatus,
		&record.ModelProvider, &record.ModelName,
		&record.PromptPath, &record.PatchPath,
		&record.StartedAt, &record.FinishedAt,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// FindAll retrieves all execution records.
func (r *ExecutionHistoryRepository) FindAll() ([]*domain.ExecutionRecord, error) {
	rows, err := r.db.Query(
		`SELECT execution_id, issue_id, execution_status, model_provider, model_name, prompt_path, patch_path, started_at, finished_at
		FROM execution_records ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []*domain.ExecutionRecord
	for rows.Next() {
		record := &domain.ExecutionRecord{}
		if err := rows.Scan(
			&record.ExecutionId, &record.IssueId, &record.ExecutionStatus,
			&record.ModelProvider, &record.ModelName,
			&record.PromptPath, &record.PatchPath,
			&record.StartedAt, &record.FinishedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}
