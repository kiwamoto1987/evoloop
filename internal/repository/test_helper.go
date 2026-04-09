package repository

import "database/sql"

// TestDB wraps a database connection for testing.
type TestDB struct {
	DB *sql.DB
}

// OpenTestDatabase creates an in-memory SQLite database for testing.
func OpenTestDatabase() (*TestDB, error) {
	db, err := OpenDatabase(":memory:")
	if err != nil {
		return nil, err
	}
	return &TestDB{DB: db}, nil
}

// Close closes the test database.
func (t *TestDB) Close() {
	_ = t.DB.Close()
}
