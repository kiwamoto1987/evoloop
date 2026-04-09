package repository

import (
	"database/sql"
	"time"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// ImprovementMemoryRepository handles persistence for ImprovementMemoryEntry.
type ImprovementMemoryRepository struct {
	db *sql.DB
}

// NewImprovementMemoryRepository creates a new repository.
func NewImprovementMemoryRepository(db *sql.DB) *ImprovementMemoryRepository {
	return &ImprovementMemoryRepository{db: db}
}

// Save inserts or replaces a memory entry.
func (r *ImprovementMemoryRepository) Save(entry *domain.ImprovementMemoryEntry) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO improvement_memory
		(memory_id, pattern_key, pattern_description, success_count, failure_count, last_observed_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		entry.MemoryId, entry.PatternKey, entry.PatternDescription,
		entry.SuccessCount, entry.FailureCount, entry.LastObservedAt,
	)
	return err
}

// FindByPatternKey retrieves a memory entry by pattern key.
func (r *ImprovementMemoryRepository) FindByPatternKey(key string) (*domain.ImprovementMemoryEntry, error) {
	row := r.db.QueryRow(
		`SELECT memory_id, pattern_key, pattern_description, success_count, failure_count, last_observed_at
		FROM improvement_memory WHERE pattern_key = ?`, key,
	)

	entry := &domain.ImprovementMemoryEntry{}
	err := row.Scan(
		&entry.MemoryId, &entry.PatternKey, &entry.PatternDescription,
		&entry.SuccessCount, &entry.FailureCount, &entry.LastObservedAt,
	)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// FindAll retrieves all memory entries.
func (r *ImprovementMemoryRepository) FindAll() ([]*domain.ImprovementMemoryEntry, error) {
	rows, err := r.db.Query(
		`SELECT memory_id, pattern_key, pattern_description, success_count, failure_count, last_observed_at
		FROM improvement_memory ORDER BY last_observed_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []*domain.ImprovementMemoryEntry
	for rows.Next() {
		entry := &domain.ImprovementMemoryEntry{}
		if err := rows.Scan(
			&entry.MemoryId, &entry.PatternKey, &entry.PatternDescription,
			&entry.SuccessCount, &entry.FailureCount, &entry.LastObservedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// RecordSuccess increments success count for a pattern key.
func (r *ImprovementMemoryRepository) RecordSuccess(patternKey string) error {
	_, err := r.db.Exec(
		`UPDATE improvement_memory SET success_count = success_count + 1, last_observed_at = ? WHERE pattern_key = ?`,
		time.Now(), patternKey,
	)
	return err
}

// RecordFailure increments failure count for a pattern key.
func (r *ImprovementMemoryRepository) RecordFailure(patternKey string) error {
	_, err := r.db.Exec(
		`UPDATE improvement_memory SET failure_count = failure_count + 1, last_observed_at = ? WHERE pattern_key = ?`,
		time.Now(), patternKey,
	)
	return err
}
