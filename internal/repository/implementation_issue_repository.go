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
		(issue_id, issue_title, issue_description, issue_category, issue_priority, issue_status, target_paths, acceptance_criteria, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		issue.IssueId,
		issue.IssueTitle,
		issue.IssueDescription,
		issue.IssueCategory,
		issue.IssuePriority,
		issue.IssueStatus,
		strings.Join(issue.TargetPaths, ","),
		strings.Join(issue.AcceptanceCriteria, ","),
		issue.CreatedAt,
	)
	return err
}

// FindByID retrieves an issue by ID.
func (r *ImplementationIssueRepository) FindByID(id string) (*domain.ImplementationIssue, error) {
	row := r.db.QueryRow(
		`SELECT issue_id, issue_title, issue_description, issue_category, issue_priority, issue_status, target_paths, acceptance_criteria, created_at
		FROM implementation_issues WHERE issue_id = ?`, id,
	)

	issue := &domain.ImplementationIssue{}
	var targetPaths, acceptanceCriteria string
	err := row.Scan(
		&issue.IssueId, &issue.IssueTitle, &issue.IssueDescription,
		&issue.IssueCategory, &issue.IssuePriority, &issue.IssueStatus,
		&targetPaths, &acceptanceCriteria, &issue.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if targetPaths != "" {
		issue.TargetPaths = strings.Split(targetPaths, ",")
	}
	if acceptanceCriteria != "" {
		issue.AcceptanceCriteria = strings.Split(acceptanceCriteria, ",")
	}

	return issue, nil
}

// FindAll retrieves all issues.
func (r *ImplementationIssueRepository) FindAll() ([]*domain.ImplementationIssue, error) {
	rows, err := r.db.Query(
		`SELECT issue_id, issue_title, issue_description, issue_category, issue_priority, issue_status, target_paths, acceptance_criteria, created_at
		FROM implementation_issues ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*domain.ImplementationIssue
	for rows.Next() {
		issue := &domain.ImplementationIssue{}
		var targetPaths, acceptanceCriteria string
		if err := rows.Scan(
			&issue.IssueId, &issue.IssueTitle, &issue.IssueDescription,
			&issue.IssueCategory, &issue.IssuePriority, &issue.IssueStatus,
			&targetPaths, &acceptanceCriteria, &issue.CreatedAt,
		); err != nil {
			return nil, err
		}
		if targetPaths != "" {
			issue.TargetPaths = strings.Split(targetPaths, ",")
		}
		if acceptanceCriteria != "" {
			issue.AcceptanceCriteria = strings.Split(acceptanceCriteria, ",")
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}
