package sqlite_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
)

// setupWatchDB creates a database with PowerSync extension loaded and hooks installed.
func setupWatchDB(t *testing.T) *sqlite.DB {
	t.Helper()

	extPath, err := powersync.ExtensionPath()
	if err != nil {
		t.Fatalf("ExtensionPath() error = %v", err)
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}

	if err := db.LoadExtension(extPath, "sqlite3_powersync_init"); err != nil {
		db.Close()
		t.Fatalf("LoadExtension() error = %v", err)
	}

	// Create a test table
	_, err = db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY, content TEXT)")
	if err != nil {
		db.Close()
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	return db
}

func TestDB_InstallUpdateHooks(t *testing.T) {
	t.Parallel()

	t.Run("installs hooks successfully", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		err := db.InstallUpdateHooks()
		if err != nil {
			t.Errorf("InstallUpdateHooks() error = %v", err)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		// Install twice
		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("first InstallUpdateHooks() error = %v", err)
		}
		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("second InstallUpdateHooks() error = %v", err)
		}

		// Should still work - insert and verify subscriber gets notified
		sub := db.Subscribe()
		defer sub.Stop()

		_, err := db.Exec("INSERT INTO messages (content) VALUES ('test')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != "messages" {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("fails without PowerSync extension", func(t *testing.T) {
		t.Parallel()

		// Open DB without loading extension
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.sqlite")
		db, err := sqlite.Open(dbPath)
		if err != nil {
			t.Fatalf("sqlite.Open() error = %v", err)
		}
		defer db.Close()

		err = db.InstallUpdateHooks()
		if err == nil {
			t.Error("expected error without PowerSync extension, got nil")
		}
	})
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

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub1 := db.Subscribe()
		defer sub1.Stop()
		sub2 := db.Subscribe()
		defer sub2.Stop()

		_, err := db.Exec("INSERT INTO messages (content) VALUES ('test')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		// Both should receive notification
		for i, sub := range []*sqlite.Subscription{sub1, sub2} {
			select {
			case tables := <-sub.Changes():
				if len(tables) != 1 || tables[0] != "messages" {
					t.Errorf("subscriber %d: expected [messages], got %v", i+1, tables)
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("subscriber %d: expected notification, got timeout", i+1)
			}
		}
	})

	t.Run("no notification without hooks installed", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		// Subscribe but don't install hooks
		sub := db.Subscribe()
		defer sub.Stop()

		_, err := db.Exec("INSERT INTO messages (content) VALUES ('test')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			t.Errorf("expected no notification without hooks, got %v", tables)
		case <-time.After(50 * time.Millisecond):
			// Expected - no notification
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

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		sub.Stop()

		_, err := db.Exec("INSERT INTO messages (content) VALUES ('test')")
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

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		_, err := db.Exec("INSERT INTO messages (content) VALUES ('hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != "messages" {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("notifies on UPDATE", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		// Insert first
		_, err := db.Exec("INSERT INTO messages (id, content) VALUES (1, 'hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		_, err = db.Exec("UPDATE messages SET content = 'world' WHERE id = 1")
		if err != nil {
			t.Fatalf("UPDATE error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != "messages" {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("notifies on DELETE", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		// Insert first
		_, err := db.Exec("INSERT INTO messages (id, content) VALUES (1, 'hello')")
		if err != nil {
			t.Fatalf("INSERT error = %v", err)
		}

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		_, err = db.Exec("DELETE FROM messages WHERE id = 1")
		if err != nil {
			t.Fatalf("DELETE error = %v", err)
		}

		select {
		case tables := <-sub.Changes():
			if len(tables) != 1 || tables[0] != "messages" {
				t.Errorf("expected [messages], got %v", tables)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("expected notification, got timeout")
		}
	})

	t.Run("no notification on SELECT", func(t *testing.T) {
		t.Parallel()

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		// SELECT doesn't go through Exec, but let's verify Query doesn't notify
		rows, err := db.Query("SELECT * FROM messages")
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

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		// This should fail - table doesn't exist
		_, err := db.Exec("INSERT INTO nonexistent (content) VALUES ('test')")
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

		db := setupWatchDB(t)
		defer db.Close()

		// Create second table
		_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("CREATE TABLE error = %v", err)
		}

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		// Insert into both tables in one transaction
		_, err = db.Exec(`
			INSERT INTO messages (content) VALUES ('hello');
			INSERT INTO users (name) VALUES ('alice');
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
			hasUsers := false
			for _, table := range tables {
				if table == "messages" {
					hasMessages = true
				}
				if table == "users" {
					hasUsers = true
				}
			}
			if !hasMessages || !hasUsers {
				t.Errorf("expected [messages, users], got %v", tables)
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

		db := setupWatchDB(t)
		defer db.Close()

		if err := db.InstallUpdateHooks(); err != nil {
			t.Fatalf("InstallUpdateHooks() error = %v", err)
		}

		sub := db.Subscribe()
		defer sub.Stop()

		// Don't read from channel - simulate slow subscriber
		// Insert multiple times
		for i := 0; i < 5; i++ {
			_, err := db.Exec("INSERT INTO messages (content) VALUES ('test')")
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
			_, _ = db.Exec("INSERT INTO messages (content) VALUES ('test')")
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
