package extension_test

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestController_StartAndLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	controller := extension.NewController(db)
	t.Cleanup(func() { _ = controller.Close() })

	instructions, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if len(instructions) == 0 {
		t.Fatal("expected at least one instruction")
	}

	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
		t.Fatalf("NotifyConnection(established) error = %v", err)
	}
	if _, err := controller.NotifyTokenRefreshed(ctx); err != nil {
		t.Fatalf("NotifyTokenRefreshed() error = %v", err)
	}
	if _, err := controller.NotifyUploadCompleted(ctx); err != nil {
		if !strings.Contains(err.Error(), "No iteration is active") {
			t.Fatalf("NotifyUploadCompleted() error = %v", err)
		}
	}
	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEnded); err != nil {
		if !strings.Contains(err.Error(), "No iteration is active") {
			t.Fatalf("NotifyConnection(end) error = %v", err)
		}
	}
}

func TestController_CheckpointFixtureReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	controller := extension.NewController(db)
	t.Cleanup(func() { _ = controller.Close() })

	if _, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
		t.Fatalf("NotifyConnection(established) error = %v", err)
	}

	path := filepath.Join("testdata", "checkpoint_lines.ndjson")
	if lineCount := replayFixture(t, controller, path); lineCount == 0 {
		t.Fatal("expected checkpoint fixture to contain lines")
	}
}

func TestController_SanitizedFixtureReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	controller := extension.NewController(db)
	t.Cleanup(func() { _ = controller.Close() })

	if _, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
		t.Fatalf("NotifyConnection(established) error = %v", err)
	}

	path := filepath.Join("testdata", "dev-sanitized.ndjson")
	if lineCount := replayFixture(t, controller, path); lineCount == 0 {
		t.Fatal("expected sanitized fixture to contain lines")
	}
}

func TestApplySchema_InitializesPowerSyncTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openBareDB(t)

	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("ApplySchema() error = %v", err)
	}

	var count int
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE name = 'ps_crud'").Scan(&count); err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Fatalf("ps_crud table missing after schema apply")
	}
}

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	ctx := context.Background()
	db := openBareDB(t)
	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("ApplySchema() error = %v", err)
	}
	return db
}

func openBareDB(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.sqlite")
	ctx := context.Background()
	db, err := sqlite.OpenBare(ctx, path)
	if err != nil {
		t.Fatalf("sqlite.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func replayFixture(t *testing.T, controller *extension.Controller, path string) int {
	t.Helper()

	ctx := context.Background()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lineNo++
		if _, err := controller.SendTextLine(ctx, line); err != nil {
			t.Fatalf("SendTextLine(line %d) error = %v", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	return lineNo
}
