package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenBareAndDBLifecycle(t *testing.T) {
	t.Parallel()

	SetExtensionPath("")

	path := filepath.Join(t.TempDir(), "tero.sqlite")
	db, err := OpenBare(context.Background(), path)
	if err != nil {
		t.Fatalf("open bare: %v", err)
	}
	if db.Path() != path {
		t.Fatalf("path mismatch: got %q want %q", db.Path(), path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file to exist: %v", err)
	}

	if _, err := db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS t(id INTEGER PRIMARY KEY, v TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(context.Background(), "INSERT INTO t(v) VALUES('ok')"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var count int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM t").Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 1 {
		t.Fatalf("count mismatch: got %d want 1", count)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := db.Exec(context.Background(), "SELECT 1"); err == nil {
		t.Fatalf("expected error after close")
	}
}

func TestOpenRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	SetExtensionPath("")
	if _, err := OpenBare(context.Background(), ""); err == nil {
		t.Fatalf("expected empty path error")
	}
}

func TestWithTxCommitsAndRollsBack(t *testing.T) {
	t.Parallel()

	SetExtensionPath("")

	db, err := OpenBare(context.Background(), filepath.Join(t.TempDir(), "tx.sqlite"))
	if err != nil {
		t.Fatalf("open bare: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS tx_items(id INTEGER PRIMARY KEY, v TEXT)"); err != nil {
		t.Fatalf("create table: %v", err)
	}

	if err := db.WithTx(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), "INSERT INTO tx_items(v) VALUES('committed')")
		return err
	}); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	var count int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tx_items").Scan(&count); err != nil {
		t.Fatalf("count after commit: %v", err)
	}
	if count != 1 {
		t.Fatalf("count after commit mismatch: got %d want 1", count)
	}

	rollbackErr := errors.New("force rollback")
	if err := db.WithTx(context.Background(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(context.Background(), "INSERT INTO tx_items(v) VALUES('rolled_back')"); err != nil {
			return err
		}
		return rollbackErr
	}); !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM tx_items").Scan(&count); err != nil {
		t.Fatalf("count after rollback: %v", err)
	}
	if count != 1 {
		t.Fatalf("count after rollback mismatch: got %d want 1", count)
	}
}

func TestWithTimeoutUsesDefaultsAndCustom(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	defaultCtx, defaultCancel := WithTimeout(ctx, 0)
	defer defaultCancel()
	defaultDeadline, ok := defaultCtx.Deadline()
	if !ok {
		t.Fatalf("expected default deadline")
	}
	defaultRemaining := time.Until(defaultDeadline)
	if defaultRemaining <= 0 || defaultRemaining > DefaultTimeout+time.Second {
		t.Fatalf("unexpected default timeout window: %v", defaultRemaining)
	}

	custom := 250 * time.Millisecond
	customCtx, customCancel := WithTimeout(ctx, custom)
	defer customCancel()
	customDeadline, ok := customCtx.Deadline()
	if !ok {
		t.Fatalf("expected custom deadline")
	}
	customRemaining := time.Until(customDeadline)
	if customRemaining <= 0 || customRemaining > custom+time.Second {
		t.Fatalf("unexpected custom timeout window: %v", customRemaining)
	}
}
