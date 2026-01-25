// Package sqlitetest provides test doubles for the sqlite package.
package sqlitetest

import (
	"database/sql"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
)

// MockDB is a test double for sqlite.Database.
type MockDB struct {
	// Counts maps table names to their row counts.
	Counts map[string]int64

	// QueryFunc is called when Query is invoked.
	QueryFunc func(query string, args ...any) (*sql.Rows, error)

	// QueryRowFunc is called when QueryRow is invoked.
	QueryRowFunc func(query string, args ...any) *sql.Row

	// ExecFunc is called when Exec is invoked.
	ExecFunc func(query string, args ...any) (sql.Result, error)

	// LoadExtensionFunc is called when LoadExtension is invoked.
	LoadExtensionFunc func(path string) error

	// Closed is set to true when Close is called.
	Closed bool
}

// Ensure MockDB implements sqlite.Database.
var _ sqlite.Database = (*MockDB)(nil)

// NewMockDB creates a new MockDB with sensible defaults.
func NewMockDB() *MockDB {
	return &MockDB{
		Counts: make(map[string]int64),
	}
}

// Query implements sqlite.Database.
func (m *MockDB) Query(query string, args ...any) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(query, args...)
	}
	return nil, fmt.Errorf("QueryFunc not set")
}

// QueryRow implements sqlite.Database.
func (m *MockDB) QueryRow(query string, args ...any) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(query, args...)
	}
	return nil
}

// Exec implements sqlite.Database.
func (m *MockDB) Exec(query string, args ...any) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(query, args...)
	}
	return nil, fmt.Errorf("ExecFunc not set")
}

// Count implements sqlite.Database.
func (m *MockDB) Count(table string) (int64, error) {
	if count, ok := m.Counts[table]; ok {
		return count, nil
	}
	return 0, nil
}

// LoadExtension implements sqlite.Database.
func (m *MockDB) LoadExtension(path string) error {
	if m.LoadExtensionFunc != nil {
		return m.LoadExtensionFunc(path)
	}
	return nil
}

// Close implements sqlite.Database.
func (m *MockDB) Close() error {
	m.Closed = true
	return nil
}
