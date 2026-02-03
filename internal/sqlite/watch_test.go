package sqlite_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/usetero/cli/internal/powersync" // registers extension via init()
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

// setupWatchDB creates a database with PowerSync extension loaded.
// Hooks are installed automatically by Open().
func setupWatchDB(t *testing.T) *sqlite.DB {
	t.Helper()

	ctx := context.Background()
	db := sqlitetest.OpenBareDB(t)

	// Create a test table
	_, err := db.Exec(ctx, "CREATE TABLE messages (id INTEGER PRIMARY KEY, content TEXT)")
	if err != nil {
		db.Close()
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	return db
}

func TestDB_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns a subscription", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		if sub == nil {
			t.Fatal("Subscribe() returned nil")
		}
		sub.Stop()
	})

	t.Run("multiple subscribers receive notifications", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub1 := db.Subscribe()
		defer sub1.Stop()
		sub2 := db.Subscribe()
		defer sub2.Stop()

		_, err := db.Exec(ctx, "INSERT INTO messages (content) VALUES ('test')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		// Both should receive notification
		for i, sub := range []*sqlite.Subscription{sub1, sub2} {
			select {
			case tables := <-sub.Changes():
				if len(tables) != 1 || tables[0] != sqlite.TableMessages {
					t.Errorf("subscriber %d: expected [messages], got %v", i+1, tables)
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("subscriber %d: expected notification, got timeout", i+1)
			}
		}
	})
}

func TestSubscription_Stop(t *testing.T) {
	t.Parallel()

	t.Run("closes the channel", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		sub.Stop()

		// Channel should be closed
		select {
		case _, ok := <-sub.Changes():
			if ok {
				t.Error("expected channel to be closed")
			}
		case <-time.After(50 * time.Millisecond):
			t.Error("expected closed channel to return immediately")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()

		// Stop multiple times should not panic
		sub.Stop()
		sub.Stop()
		sub.Stop()
	})

	t.Run("stops receiving notifications", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		sub.Stop()

		_, err := db.Exec(ctx, "INSERT INTO messages (content) VALUES ('test')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		// Should not receive notification (channel closed)
		select {
		case _, ok := <-sub.Changes():
			if ok {
				t.Error("expected channel to be closed, got value")
			}
		case <-time.After(50 * time.Millisecond):
			t.Error("expected closed channel to return immediately")
		}
	})
}

func TestDB_Exec_NotifiesSubscribers(t *testing.T) {
	t.Parallel()

	t.Run("notifies on INSERT", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		defer sub.Stop()

		_, err := db.Exec(ctx, "INSERT INTO messages (content) VALUES ('hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != sqlite.TableMessages {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("notifies on UPDATE", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		// Insert first
		_, err := db.Exec(ctx, "INSERT INTO messages (id, content) VALUES (1, 'hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		// Drain the INSERT notification
		sub := db.Subscribe()
		defer sub.Stop()

		_, err = db.Exec(ctx, "UPDATE messages SET content = 'world' WHERE id = 1")
		if err != nil {
			t.Fatalf("UPDATE error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != sqlite.TableMessages {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("notifies on DELETE", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		// Insert first
		_, err := db.Exec(ctx, "INSERT INTO messages (id, content) VALUES (1, 'hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		_, err = db.Exec(ctx, "DELETE FROM messages WHERE id = 1")
		if err != nil {
			t.Fatalf("DELETE error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != sqlite.TableMessages {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("no notification on SELECT", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		defer sub.Stop()

		// SELECT doesn't go through Exec, but let's verify Query doesn't notify
		rows, err := db.Query(ctx, "SELECT * FROM messages")
		if err != nil {
			t.Fatalf("SELECT error = %v", err)
		}
		rows.Close()

		select {
		case tables := <-sub.Changes():
			t.Errorf("expected no notification on SELECT, got %v", tables)
		case <-time.After(50 * time.Millisecond):
			// Expected - no notification
		}
	})

	t.Run("no notification on failed Exec", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		defer sub.Stop()

		// This should fail - table doesn't exist
		_, err := db.Exec(ctx, "INSERT INTO nonexistent (content) VALUES ('test')")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		select {
		case tables := <-sub.Changes():
			t.Errorf("expected no notification on failed exec, got %v", tables)
		case <-time.After(50 * time.Millisecond):
			// Expected - no notification
		}
	})

	t.Run("notifies with multiple tables changed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)

		sub := db.Subscribe()
		defer sub.Stop()

		// Insert into both messages and conversations
		_, err := db.Exec(ctx, `
			INSERT INTO messages (id, account_id, conversation_id, role) VALUES ('msg-1', 'acc-1', 'conv-1', 'user');
			INSERT INTO conversations (id, account_id, title) VALUES ('conv-1', 'acc-1', 'Test');
		`)
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 2 {
				t.Errorf("expected 2 tables, got %v", tables)
			}
			// Check both tables are present (order may vary)
			hasMessages := false
			hasConversations := false
			for _, table := range tables {
				if table == sqlite.TableMessages {
					hasMessages = true
				}
				if table == sqlite.TableConversations {
					hasConversations = true
				}
			}
			if !hasMessages || !hasConversations {
				t.Errorf("expected [messages, conversations], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})
}

func TestDB_Messages_NotifiesSubscribers(t *testing.T) {
	t.Parallel()

	t.Run("CreateAssistantMessage notifies subscribers", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)

		sub := db.Subscribe()
		defer sub.Stop()

		_, err := db.Messages().CreateAssistantMessage(ctx, "acc-1", "conv-1", "claude-3")
		if err != nil {
			t.Fatalf("CreateAssistantMessage() error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			hasMessages := false
			for _, table := range tables {
				if table == sqlite.TableMessages {
					hasMessages = true
					break
				}
			}
			if !hasMessages {
				t.Errorf("expected messages table in %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("UpdateContent notifies subscribers", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)

		// Create message first
		msgID, err := db.Messages().CreateAssistantMessage(ctx, "acc-1", "conv-1", "claude-3")
		if err != nil {
			t.Fatalf("CreateAssistantMessage() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		err = db.Messages().UpdateContent(ctx, msgID, `[{"type":"text"}]`)
		if err != nil {
			t.Fatalf("UpdateContent() error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			hasMessages := false
			for _, table := range tables {
				if table == sqlite.TableMessages {
					hasMessages = true
					break
				}
			}
			if !hasMessages {
				t.Errorf("expected messages table in %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("UpdateMeta notifies subscribers", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := dbtest.OpenTestDB(t)

		// Create message first
		msgID, err := db.Messages().CreateAssistantMessage(ctx, "acc-1", "conv-1", "claude-3")
		if err != nil {
			t.Fatalf("CreateAssistantMessage() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		err = db.Messages().UpdateMeta(ctx, msgID, "claude-3", "end_turn")
		if err != nil {
			t.Fatalf("UpdateMeta() error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			hasMessages := false
			for _, table := range tables {
				if table == sqlite.TableMessages {
					hasMessages = true
					break
				}
			}
			if !hasMessages {
				t.Errorf("expected messages table in %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})
}

func TestDB_Subscribe_BufferBehavior(t *testing.T) {
	t.Parallel()

	t.Run("drops notifications when subscriber is slow", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db := setupWatchDB(t)
		defer db.Close()

		sub := db.Subscribe()
		defer sub.Stop()

		// Don't read from channel - simulate slow subscriber
		// Insert multiple times
		for i := 0; i < 5; i++ {
			_, err := db.Exec(ctx, "INSERT INTO messages (content) VALUES ('test')")
			if err != nil {
				t.Fatalf("INSERT error = %v", err)
			}
		}

		// Should get at least one notification (buffer size is 1)
		select {
		case <-sub.Changes():
			// Good - got at least one
		case <-time.After(100 * time.Millisecond):
			t.Error("expected at least one notification")
		}

		// Subsequent writes should not block (non-blocking send)
		done := make(chan bool)
		go func() {
			_, _ = db.Exec(ctx, "INSERT INTO messages (content) VALUES ('test')")
			done <- true
		}()

		select {
		case <-done:
			// Good - didn't block
		case <-time.After(100 * time.Millisecond):
			t.Error("Exec blocked on slow subscriber")
		}
	})
}
