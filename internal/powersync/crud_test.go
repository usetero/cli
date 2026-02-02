package powersync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
)

func TestCrudQueue_GetNextEntry(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when queue is empty", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		queue := powersync.NewCrudQueue(db)

		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}
		if entry != nil {
			t.Errorf("GetNextEntry() = %v, want nil", entry)
		}
	})

	t.Run("returns entry with parsed data", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{"content":"hello"}}`)

		queue := powersync.NewCrudQueue(db)
		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}

		if entry == nil {
			t.Fatal("GetNextEntry() = nil, want entry")
		}
		if entry.ID != 1 {
			t.Errorf("entry.ID = %d, want 1", entry.ID)
		}
		if entry.Op != "PUT" {
			t.Errorf("entry.Op = %q, want PUT", entry.Op)
		}
		if entry.Table != "messages" {
			t.Errorf("entry.Table = %q, want messages", entry.Table)
		}
		if entry.RowID != "msg-1" {
			t.Errorf("entry.RowID = %q, want msg-1", entry.RowID)
		}
		if entry.Data["content"] != "hello" {
			t.Errorf("entry.Data[content] = %v, want hello", entry.Data["content"])
		}
	})

	t.Run("returns entries in order", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"first","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 2, nil, `{"op":"PUT","type":"messages","id":"second","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		entry, _ := queue.GetNextEntry(context.Background())
		if entry.RowID != "first" {
			t.Errorf("first entry.RowID = %q, want first", entry.RowID)
		}
	})

	t.Run("returns error on malformed JSON", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `not valid json`)

		queue := powersync.NewCrudQueue(db)

		_, err := queue.GetNextEntry(context.Background())
		if err == nil {
			t.Error("GetNextEntry() should return error for malformed JSON")
		}
	})
}

func TestCrudQueue_GetAllEntries(t *testing.T) {
	t.Parallel()

	t.Run("returns empty slice when queue is empty", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		queue := powersync.NewCrudQueue(db)

		entries, err := queue.GetAllEntries(context.Background())
		if err != nil {
			t.Fatalf("GetAllEntries() error = %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("GetAllEntries() = %d entries, want 0", len(entries))
		}
	})

	t.Run("returns all entries in order", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"first","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 2, nil, `{"op":"PATCH","type":"messages","id":"second","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 3, nil, `{"op":"DELETE","type":"messages","id":"third","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		entries, err := queue.GetAllEntries(context.Background())
		if err != nil {
			t.Fatalf("GetAllEntries() error = %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("GetAllEntries() = %d entries, want 3", len(entries))
		}

		if entries[0].RowID != "first" {
			t.Errorf("entries[0].RowID = %q, want first", entries[0].RowID)
		}
		if entries[1].RowID != "second" {
			t.Errorf("entries[1].RowID = %q, want second", entries[1].RowID)
		}
		if entries[2].RowID != "third" {
			t.Errorf("entries[2].RowID = %q, want third", entries[2].RowID)
		}
	})
}

func TestCrudQueue_HasPendingUploads(t *testing.T) {
	t.Parallel()

	t.Run("returns false when empty", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		queue := powersync.NewCrudQueue(db)

		has, err := queue.HasPendingUploads(context.Background())
		if err != nil {
			t.Fatalf("HasPendingUploads() error = %v", err)
		}
		if has {
			t.Error("HasPendingUploads() = true, want false")
		}
	})

	t.Run("returns true when entries exist", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		has, err := queue.HasPendingUploads(context.Background())
		if err != nil {
			t.Fatalf("HasPendingUploads() error = %v", err)
		}
		if !has {
			t.Error("HasPendingUploads() = false, want true")
		}
	})
}

func TestCrudQueue_CheckHealth(t *testing.T) {
	t.Parallel()

	t.Run("healthy database passes", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		queue := powersync.NewCrudQueue(db)
		err := queue.CheckHealth(ctx)
		if err != nil {
			t.Fatalf("CheckHealth() error = %v", err)
		}
	})

	t.Run("missing ps_tx row returns corrupt error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Delete the required row
		_, err := db.Exec(ctx, "DELETE FROM ps_tx")
		if err != nil {
			t.Fatalf("delete ps_tx: %v", err)
		}

		queue := powersync.NewCrudQueue(db)
		err = queue.CheckHealth(ctx)
		if err == nil {
			t.Fatal("CheckHealth() should return error")
		}
		if !errors.Is(err, powersync.ErrDatabaseCorrupt) {
			t.Errorf("CheckHealth() error = %v, want ErrDatabaseCorrupt", err)
		}
	})

	t.Run("stuck local bucket returns corrupt error", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Set up $local bucket in stuck state (target_op > last_op with empty ps_crud)
		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 5)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		queue := powersync.NewCrudQueue(db)
		err = queue.CheckHealth(ctx)
		if err == nil {
			t.Fatal("CheckHealth() should return error")
		}
		if !errors.Is(err, powersync.ErrDatabaseCorrupt) {
			t.Errorf("CheckHealth() error = %v, want ErrDatabaseCorrupt", err)
		}
	})

	t.Run("local bucket with pending data is healthy", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		// Set up $local bucket with target_op > last_op, but with actual pending data
		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 5)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)

		queue := powersync.NewCrudQueue(db)
		err = queue.CheckHealth(ctx)
		if err != nil {
			t.Fatalf("CheckHealth() error = %v", err)
		}
	})

	t.Run("no local bucket is healthy", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		queue := powersync.NewCrudQueue(db)
		err := queue.CheckHealth(ctx)
		if err != nil {
			t.Fatalf("CheckHealth() error = %v", err)
		}
	})
}

func TestCrudQueue_GetNextTransaction(t *testing.T) {
	t.Parallel()

	t.Run("returns single entry when no tx_id", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 2, nil, `{"op":"PUT","type":"messages","id":"msg-2","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		entries, err := queue.GetNextTransaction(context.Background())
		if err != nil {
			t.Fatalf("GetNextTransaction() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("GetNextTransaction() returned %d entries, want 1", len(entries))
		}
	})

	t.Run("returns all entries with same tx_id", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		txID := int64(100)
		powersynctest.InsertCrudEntry(t, db, 1, &txID, `{"op":"PUT","type":"conversations","id":"conv-1","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 2, &txID, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)
		powersynctest.InsertCrudEntry(t, db, 3, nil, `{"op":"PUT","type":"messages","id":"msg-2","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		entries, err := queue.GetNextTransaction(context.Background())
		if err != nil {
			t.Fatalf("GetNextTransaction() error = %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("GetNextTransaction() returned %d entries, want 2", len(entries))
		}
	})
}
