package repository

import (
	"database/sql"
	"strings"

	"github.com/kiwamoto1987/evoloop/internal/domain"
)

// ImplementationIssueRepository handles persistence for ImplementationIssue.
type ImplementationIssueRepository struct {
	db *sql.DB
}

// NewImplementationIssueRepository creates a new repository.
func NewImplementationIssueRepository(db *sql.DB) *ImplementationIssueRepository {
	return &ImplementationIssueRepository{db: db}
}

// Save inserts or replaces an issue.
func (r *ImplementationIssueRepository) Save(issue *domain.ImplementationIssue) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO implementation_issues
		(issue_id, issue_title, issue_description, issue_category, remediation_type, issue_priority, issue_status, target_paths, acceptance_criteria, source, source_ref, dedup_key, attempt_count, last_attempted_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		issue.IssueId,
		issue.IssueTitle,
		issue.IssueDescription,
		issue.IssueCategory,
		issue.RemediationType,
		issue.IssuePriority,
		issue.IssueStatus,
		strings.Join(issue.TargetPaths, ","),
		strings.Join(issue.AcceptanceCriteria, ","),
		issue.Source,
		issue.SourceRef,
		issue.DedupKey,
		issue.AttemptCount,
		issue.LastAttemptedAt,
		issue.CreatedAt,
	)
	return err
}

// FindByID retrieves an issue by ID.
func (r *ImplementationIssueRepository) FindByID(id string) (*domain.ImplementationIssue, error) {
	row := r.db.QueryRow(
		`SELECT issue_id, issue_title, issue_description, issue_category, remediation_type, issue_priority, issue_status, target_paths, acceptance_criteria, source, source_ref, dedup_key, attempt_count, last_attempted_at, created_at
		FROM implementation_issues WHERE issue_id = ?`, id,
	)
	return scanIssue(row)
}

// FindAll retrieves all issues.
func (r *ImplementationIssueRepository) FindAll() ([]*domain.ImplementationIssue, error) {
	rows, err := r.db.Query(
		`SELECT issue_id, issue_title, issue_description, issue_category, remediation_type, issue_priority, issue_status, target_paths, acceptance_criteria, source, source_ref, dedup_key, attempt_count, last_attempted_at, created_at
		FROM implementation_issues ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanIssues(rows)
}

// FindOpenProposable returns Open issues that are proposable, ordered by priority.
func (r *ImplementationIssueRepository) FindOpenProposable() ([]*domain.ImplementationIssue, error) {
	rows, err := r.db.Query(
		`SELECT issue_id, issue_title, issue_description, issue_category, remediation_type, issue_priority, issue_status, target_paths, acceptance_criteria, source, source_ref, dedup_key, attempt_count, last_attempted_at, created_at
		FROM implementation_issues
		WHERE issue_status = ? AND issue_category != ?
		ORDER BY issue_priority ASC, attempt_count ASC`,
		domain.IssueStatusOpen, domain.IssueCategoryEnvironment,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanIssues(rows)
}

// FindByDedupKey retrieves an Open issue by dedup key. Returns nil if not found.
func (r *ImplementationIssueRepository) FindByDedupKey(key string) (*domain.ImplementationIssue, error) {
	row := r.db.QueryRow(
		`SELECT issue_id, issue_title, issue_description, issue_category, remediation_type, issue_priority, issue_status, target_paths, acceptance_criteria, source, source_ref, dedup_key, attempt_count, last_attempted_at, created_at
		FROM implementation_issues
		WHERE dedup_key = ? AND issue_status = ?`,
		key, domain.IssueStatusOpen,
	)

	issue, err := scanIssue(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return issue, err
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanIssue(row scannable) (*domain.ImplementationIssue, error) {
	issue := &domain.ImplementationIssue{}
	var targetPaths, acceptanceCriteria string
	var dedupKey sql.NullString
	err := row.Scan(
		&issue.IssueId, &issue.IssueTitle, &issue.IssueDescription,
		&issue.IssueCategory, &issue.RemediationType,
		&issue.IssuePriority, &issue.IssueStatus,
		&targetPaths, &acceptanceCriteria,
		&issue.Source, &issue.SourceRef, &dedupKey,
		&issue.AttemptCount, &issue.LastAttemptedAt, &issue.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if dedupKey.Valid {
		issue.DedupKey = dedupKey.String
	}
	if targetPaths != "" {
		issue.TargetPaths = strings.Split(targetPaths, ",")
	}
	if acceptanceCriteria != "" {
		issue.AcceptanceCriteria = strings.Split(acceptanceCriteria, ",")
	}

	return issue, nil
}

func scanIssues(rows *sql.Rows) ([]*domain.ImplementationIssue, error) {
	var issues []*domain.ImplementationIssue
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}
