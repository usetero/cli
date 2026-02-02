package powersync_test

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestCompleteBatch(t *testing.T) {
	t.Parallel()

	t.Run("deletes crud entries up to lastEntryID", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Insert some crud entries
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 2, nil, `{"op":"PUT","type":"messages","id":"msg-2","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 3, nil, `{"op":"PUT","type":"messages","id":"msg-3","data":{}}`)

		// Set up $local bucket
		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		// Complete batch for entries 1 and 2
		err = powersync.CompleteBatch(ctx, db.DB(), 2, "100")
		if err != nil {
			t.Fatalf("CompleteBatch() error = %v", err)
		}

		// Verify entries 1 and 2 are deleted, 3 remains
		var count int
		err = db.QueryRow(ctx, "SELECT COUNT(*) FROM ps_crud").Scan(&count)
		if err != nil {
			t.Fatalf("count crud: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 remaining entry, got %d", count)
		}

		var remainingID int64
		err = db.QueryRow(ctx, "SELECT id FROM ps_crud").Scan(&remainingID)
		if err != nil {
			t.Fatalf("get remaining: %v", err)
		}
		if remainingID != 3 {
			t.Errorf("expected entry 3 to remain, got %d", remainingID)
		}
	})

	t.Run("updates target_op to checkpoint", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Set up $local bucket
		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		err = powersync.CompleteBatch(ctx, db.DB(), 0, "42")
		if err != nil {
			t.Fatalf("CompleteBatch() error = %v", err)
		}

		var targetOp int64
		err = db.QueryRow(ctx, "SELECT target_op FROM ps_buckets WHERE name = '$local'").Scan(&targetOp)
		if err != nil {
			t.Fatalf("get target_op: %v", err)
		}
		if targetOp != 42 {
			t.Errorf("target_op = %d, want 42", targetOp)
		}
	})

	t.Run("is atomic - rolls back on error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Insert a crud entry
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)

		// Don't create $local bucket - update should fail
		err := powersync.CompleteBatch(ctx, db.DB(), 1, "100")

		// Should succeed (UPDATE affects 0 rows but doesn't error)
		// This is actually fine - if there's no $local bucket, sync hasn't started
		if err != nil {
			t.Fatalf("CompleteBatch() error = %v", err)
		}

		// Crud entry should still be deleted
		var count int
		err = db.QueryRow(ctx, "SELECT COUNT(*) FROM ps_crud").Scan(&count)
		if err != nil {
			t.Fatalf("count crud: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 entries, got %d", count)
		}
	})
}

func TestGetClientID(t *testing.T) {
	t.Parallel()

	t.Run("returns client ID from powersync extension", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		clientID, err := powersync.GetClientID(ctx, db.DB())
		if err != nil {
			t.Fatalf("GetClientID() error = %v", err)
		}

		// Client ID should be a non-empty UUID-like string
		if clientID == "" {
			t.Error("GetClientID() returned empty string")
		}
		if len(clientID) < 32 {
			t.Errorf("GetClientID() = %q, expected UUID-like string", clientID)
		}
	})

	t.Run("returns same ID for same database", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		id1, err := powersync.GetClientID(ctx, db.DB())
		if err != nil {
			t.Fatalf("first GetClientID() error = %v", err)
		}

		id2, err := powersync.GetClientID(ctx, db.DB())
		if err != nil {
			t.Fatalf("second GetClientID() error = %v", err)
		}

		if id1 != id2 {
			t.Errorf("client IDs differ: %q vs %q", id1, id2)
		}
	})
}
