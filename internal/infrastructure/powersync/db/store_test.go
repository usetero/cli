package db_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestStoreNextMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty queue returns nil", func(t *testing.T) {
		t.Parallel()
		store := db.NewStore(openTestDB(t))
		got, err := store.NextMutation(ctx)
		if err != nil {
			t.Fatalf("NextMutation() error = %v", err)
		}
		if got != nil {
			t.Fatalf("NextMutation() = %v, want nil", got)
		}
	})

	t.Run("parses mutation", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","id":"msg-1","data":{"content":"hello"}}`)

		store := db.NewStore(database)
		got, err := store.NextMutation(ctx)
		if err != nil {
			t.Fatalf("NextMutation() error = %v", err)
		}
		if got == nil {
			t.Fatal("NextMutation() = nil")
		}
		if got.ID != 1 || got.Op != db.OperationPut || got.Table != db.TableMessages || got.RowID != "msg-1" {
			t.Fatalf("unexpected mutation: %+v", *got)
		}
		if got.Data["content"] != "hello" {
			t.Fatalf("data content = %v", got.Data["content"])
		}
	})

	t.Run("malformed data returns error", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `not json`)
		store := db.NewStore(database)
		if _, err := store.NextMutation(ctx); err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("invalid op returns error", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `{"op":"UPSERT","type":"messages","id":"msg-1","data":{}}`)
		store := db.NewStore(database)
		if _, err := store.NextMutation(ctx); err == nil {
			t.Fatal("expected invalid op error")
		}
	})

	t.Run("missing table type returns error", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `{"op":"PUT","id":"msg-1","data":{}}`)
		store := db.NewStore(database)
		if _, err := store.NextMutation(ctx); err == nil {
			t.Fatal("expected missing type error")
		}
	})

	t.Run("missing row id returns error", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","data":{}}`)
		store := db.NewStore(database)
		if _, err := store.NextMutation(ctx); err == nil {
			t.Fatal("expected missing id error")
		}
	})
}

func TestStoreNextMutationBatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns single entry when tx_id is null", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{}}`)
		insertCrud(t, database, 2, nil, `{"op":"PUT","type":"messages","id":"m2","data":{}}`)

		store := db.NewStore(database)
		batch, err := store.NextMutationBatch(ctx)
		if err != nil {
			t.Fatalf("NextMutationBatch() error = %v", err)
		}
		if len(batch) != 1 || batch[0].RowID != "m1" {
			t.Fatalf("batch = %+v", batch)
		}
	})

	t.Run("returns transaction group when tx_id present", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		txID := int64(99)
		insertCrud(t, database, 1, &txID, `{"op":"PUT","type":"messages","id":"m1","data":{}}`)
		insertCrud(t, database, 2, &txID, `{"op":"PATCH","type":"messages","id":"m2","data":{}}`)
		insertCrud(t, database, 3, nil, `{"op":"DELETE","type":"messages","id":"m3"}`)

		store := db.NewStore(database)
		batch, err := store.NextMutationBatch(ctx)
		if err != nil {
			t.Fatalf("NextMutationBatch() error = %v", err)
		}
		if len(batch) != 2 || batch[0].RowID != "m1" || batch[1].RowID != "m2" {
			t.Fatalf("batch = %+v", batch)
		}
	})
}

func TestStorePendingAndHasPending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database := openTestDB(t)
	store := db.NewStore(database)

	has, err := store.HasPendingMutations(ctx)
	if err != nil {
		t.Fatalf("HasPendingMutations() error = %v", err)
	}
	if has {
		t.Fatal("expected no pending mutations")
	}

	insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{}}`)
	insertCrud(t, database, 2, nil, `{"op":"PATCH","type":"messages","id":"m2","data":{}}`)

	has, err = store.HasPendingMutations(ctx)
	if err != nil {
		t.Fatalf("HasPendingMutations() error = %v", err)
	}
	if !has {
		t.Fatal("expected pending mutations")
	}

	pending, err := store.PendingMutations(ctx)
	if err != nil {
		t.Fatalf("PendingMutations() error = %v", err)
	}
	if len(pending) != 2 || pending[0].RowID != "m1" || pending[1].RowID != "m2" {
		t.Fatalf("pending = %+v", pending)
	}
}

func TestStoreCompleteUploadedBatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database := openTestDB(t)
	store := db.NewStore(database)

	insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{}}`)
	insertCrud(t, database, 2, nil, `{"op":"PUT","type":"messages","id":"m2","data":{}}`)
	insertCrud(t, database, 3, nil, `{"op":"PUT","type":"messages","id":"m3","data":{}}`)

	if _, err := database.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES (?, 0, 0)", string(db.LocalBucket)); err != nil {
		t.Fatalf("seed local bucket: %v", err)
	}

	if err := store.CompleteUploadedBatch(ctx, 2, db.OpID(42)); err != nil {
		t.Fatalf("CompleteUploadedBatch() error = %v", err)
	}

	var count int
	if err := database.QueryRow(ctx, "SELECT COUNT(*) FROM ps_crud").Scan(&count); err != nil {
		t.Fatalf("count ps_crud: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	var targetOp int64
	if err := database.QueryRow(ctx, "SELECT target_op FROM ps_buckets WHERE name = ?", string(db.LocalBucket)).Scan(&targetOp); err != nil {
		t.Fatalf("read target_op: %v", err)
	}
	if targetOp != 42 {
		t.Fatalf("target_op = %d, want 42", targetOp)
	}
}

func TestStoreClientID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := db.NewStore(openTestDB(t))

	id1, err := store.ClientID(ctx)
	if err != nil {
		t.Fatalf("ClientID() error = %v", err)
	}
	id2, err := store.ClientID(ctx)
	if err != nil {
		t.Fatalf("ClientID() second error = %v", err)
	}
	if id1 == "" || id2 == "" || id1 != id2 {
		t.Fatalf("client IDs invalid: %q %q", id1, id2)
	}
}

func TestStoreCheckHealth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("healthy passes", func(t *testing.T) {
		t.Parallel()
		store := db.NewStore(openTestDB(t))
		if err := store.CheckHealth(ctx); err != nil {
			t.Fatalf("CheckHealth() error = %v", err)
		}
	})

	t.Run("missing ps_tx row is corrupt", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		if _, err := database.Exec(ctx, "DELETE FROM ps_tx"); err != nil {
			t.Fatalf("delete ps_tx: %v", err)
		}
		store := db.NewStore(database)
		err := store.CheckHealth(ctx)
		if !errors.Is(err, db.ErrCorrupt) {
			t.Fatalf("expected ErrCorrupt, got %v", err)
		}
	})

	t.Run("stuck local bucket is corrupt when queue empty", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		if _, err := database.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES (?, 0, 5)", string(db.LocalBucket)); err != nil {
			t.Fatalf("insert local bucket: %v", err)
		}
		store := db.NewStore(database)
		err := store.CheckHealth(ctx)
		if !errors.Is(err, db.ErrCorrupt) {
			t.Fatalf("expected ErrCorrupt, got %v", err)
		}
	})

	t.Run("stuck local bucket is tolerated when queue has pending", func(t *testing.T) {
		t.Parallel()
		database := openTestDB(t)
		if _, err := database.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES (?, 0, 5)", string(db.LocalBucket)); err != nil {
			t.Fatalf("insert local bucket: %v", err)
		}
		insertCrud(t, database, 1, nil, `{"op":"PUT","type":"messages","id":"m1","data":{}}`)
		store := db.NewStore(database)
		if err := store.CheckHealth(ctx); err != nil {
			t.Fatalf("CheckHealth() error = %v", err)
		}
	})
}

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("extension.Register() error = %v", err)
	}
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "powersync-db-test.sqlite")
	database, err := sqlite.OpenBare(ctx, path)
	if err != nil {
		t.Fatalf("sqlite.OpenBare() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := extension.ApplySchema(ctx, database); err != nil {
		t.Fatalf("extension.ApplySchema() error = %v", err)
	}
	return database
}

func insertCrud(t *testing.T, database *sqlite.DB, id int64, txID *int64, data string) {
	t.Helper()
	ctx := context.Background()
	if _, err := database.Exec(ctx, "INSERT INTO ps_crud (id, tx_id, data) VALUES (?, ?, ?)", id, txID, data); err != nil {
		t.Fatalf("insert crud row: %v", err)
	}
}
