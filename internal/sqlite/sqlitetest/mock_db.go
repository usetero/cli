// Package sqlitetest provides test doubles for the sqlite package.
package sqlitetest

import (
	"context"
	"database/sql"

	"github.com/usetero/cli/internal/sqlite"
)

// MockDB is a test double for sqlite.DB.
type MockDB struct {
	// MessagesImpl is the mock messages implementation.
	MessagesImpl sqlite.Messages

	// ConversationsImpl is the mock conversations implementation.
	ConversationsImpl sqlite.Conversations

	// DatadogAccountStatusesImpl is the mock Datadog account statuses implementation.
	DatadogAccountStatusesImpl sqlite.DatadogAccountStatuses

	// ServiceStatusesImpl is the mock service statuses implementation.
	ServiceStatusesImpl sqlite.ServiceStatuses

	// ServicesImpl is the mock services implementation.
	ServicesImpl sqlite.Services

	// LogEventsImpl is the mock log events implementation.
	LogEventsImpl sqlite.LogEvents

	// LogEventPoliciesImpl is the mock log event policies implementation.
	LogEventPoliciesImpl sqlite.LogEventPolicies

	// QueryFunc is called by Query.
	QueryFunc func(ctx context.Context, sql string, args ...any) (*sql.Rows, error)

	// QueryRowFunc is called by QueryRow.
	QueryRowFunc func(ctx context.Context, sql string, args ...any) *sql.Row

	// ExecFunc is called by Exec.
	ExecFunc func(ctx context.Context, sql string, args ...any) (sql.Result, error)

	// Closed is set to true when Close is called.
	Closed bool
}

// Ensure MockDB implements sqlite.DB.
var _ sqlite.DB = (*MockDB)(nil)

// NewMockDB creates a new MockDB with sensible defaults.
func NewMockDB() *MockDB {
	return &MockDB{}
}

// Messages implements sqlite.DB.
func (m *MockDB) Messages() sqlite.Messages {
	return m.MessagesImpl
}

// Conversations implements sqlite.DB.
func (m *MockDB) Conversations() sqlite.Conversations {
	return m.ConversationsImpl
}

// DatadogAccountStatuses implements sqlite.DB.
func (m *MockDB) DatadogAccountStatuses() sqlite.DatadogAccountStatuses {
	return m.DatadogAccountStatusesImpl
}

// ServiceStatuses implements sqlite.DB.
func (m *MockDB) ServiceStatuses() sqlite.ServiceStatuses {
	return m.ServiceStatusesImpl
}

// Query implements sqlite.DB.
func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

// QueryRow implements sqlite.DB.
func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

// Exec implements sqlite.DB.
func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return nil, nil
}

// WithTx implements sqlite.DB.
func (m *MockDB) WithTx(ctx context.Context, fn func(tx *sqlite.Tx) error) error {
	return nil
}

// Raw implements sqlite.DB.
func (m *MockDB) Raw() *sql.DB {
	return nil
}

// ReadRaw implements sqlite.DB.
func (m *MockDB) ReadRaw() *sql.DB {
	return nil
}

// Services implements sqlite.DB.
func (m *MockDB) Services() sqlite.Services {
	return m.ServicesImpl
}

// LogEvents implements sqlite.DB.
func (m *MockDB) LogEvents() sqlite.LogEvents {
	return m.LogEventsImpl
}

// LogEventPolicies implements sqlite.DB.
func (m *MockDB) LogEventPolicies() sqlite.LogEventPolicies {
	return m.LogEventPoliciesImpl
}

// PendingUploadCounts implements sqlite.DB.
func (m *MockDB) PendingUploadCounts(_ context.Context) (map[sqlite.Table]int64, error) {
	return nil, nil
}

// Close implements sqlite.DB.
func (m *MockDB) Close() error {
	m.Closed = true
	return nil
}
