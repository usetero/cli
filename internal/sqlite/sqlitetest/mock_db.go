// Package sqlitetest provides test doubles for the sqlite package.
package sqlitetest

import (
	"github.com/usetero/cli/internal/sqlite"
)

// MockDB is a test double for sqlite.Database.
type MockDB struct {
	// MessagesImpl is the mock messages implementation.
	MessagesImpl sqlite.Messages

	// ConversationsImpl is the mock conversations implementation.
	ConversationsImpl sqlite.Conversations

	// SubscriptionImpl is returned by Subscribe.
	SubscriptionImpl *sqlite.Subscription

	// DBImpl is returned by DB() for low-level access.
	DBImpl *sqlite.DB

	// Closed is set to true when Close is called.
	Closed bool
}

// Ensure MockDB implements sqlite.Database.
var _ sqlite.Database = (*MockDB)(nil)

// NewMockDB creates a new MockDB with sensible defaults.
func NewMockDB() *MockDB {
	return &MockDB{}
}

// Messages implements sqlite.Database.
func (m *MockDB) Messages() sqlite.Messages {
	return m.MessagesImpl
}

// Conversations implements sqlite.Database.
func (m *MockDB) Conversations() sqlite.Conversations {
	return m.ConversationsImpl
}

// Subscribe implements sqlite.Database.
func (m *MockDB) Subscribe() *sqlite.Subscription {
	return m.SubscriptionImpl
}

// Close implements sqlite.Database.
func (m *MockDB) Close() error {
	m.Closed = true
	return nil
}

// DB implements sqlite.Database.
func (m *MockDB) DB() *sqlite.DB {
	return m.DBImpl
}
