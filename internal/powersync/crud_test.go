package powersync_test

import (
	"context"
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

func TestCrudQueue_DeleteEntry(t *testing.T) {
	t.Parallel()

	t.Run("removes entry from queue", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{}}`)

		queue := powersync.NewCrudQueue(db)

		err := queue.DeleteEntry(context.Background(), 1)
		if err != nil {
			t.Fatalf("DeleteEntry() error = %v", err)
		}

		entry, _ := queue.GetNextEntry(context.Background())
		if entry != nil {
			t.Errorf("entry still exists after delete")
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
