package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDatabase opens or creates the SQLite database and runs migrations.
func OpenDatabase(dbPath string) (*sql.DB, error) {
	// Ensure parent directory exists
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS implementation_issues (
			issue_id TEXT PRIMARY KEY,
			issue_title TEXT NOT NULL,
			issue_description TEXT NOT NULL,
			issue_category TEXT NOT NULL,
			issue_priority INTEGER NOT NULL,
			issue_status TEXT NOT NULL,
			target_paths TEXT,
			acceptance_criteria TEXT,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS execution_records (
			execution_id TEXT PRIMARY KEY,
			issue_id TEXT NOT NULL,
			execution_status TEXT NOT NULL,
			model_provider TEXT,
			model_name TEXT,
			prompt_path TEXT,
			patch_path TEXT,
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			FOREIGN KEY (issue_id) REFERENCES implementation_issues(issue_id)
		)`,
		`CREATE TABLE IF NOT EXISTS evaluation_reports (
			evaluation_id TEXT PRIMARY KEY,
			execution_id TEXT NOT NULL,
			test_passed BOOLEAN NOT NULL,
			lint_passed BOOLEAN NOT NULL,
			typecheck_passed BOOLEAN NOT NULL,
			changed_file_count INTEGER NOT NULL,
			changed_line_count INTEGER NOT NULL,
			evaluation_decision TEXT NOT NULL,
			failure_reasons TEXT,
			generated_at DATETIME NOT NULL,
			FOREIGN KEY (execution_id) REFERENCES execution_records(execution_id)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w\nstatement: %s", err, stmt)
		}
	}

	return nil
}
