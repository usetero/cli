package integration_test

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
	"github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/powersync/uploader"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type pipelineClient struct {
	mu sync.Mutex

	token      psclient.AccessToken
	checkpoint psclient.WriteCheckpoint
	streamLine []byte

	writeCheckpointCalls atomic.Int32
}

func (c *pipelineClient) SetToken(token psclient.AccessToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

func (c *pipelineClient) SyncStream(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
	if len(c.streamLine) > 0 {
		line := make([]byte, len(c.streamLine))
		copy(line, c.streamLine)
		if err := handler(line); err != nil {
			return err
		}
	}
	<-ctx.Done()
	return ctx.Err()
}

func (c *pipelineClient) GetWriteCheckpoint(context.Context, psclient.ClientID) (psclient.WriteCheckpoint, error) {
	c.writeCheckpointCalls.Add(1)
	if c.checkpoint == "" {
		return "0", nil
	}
	return c.checkpoint, nil
}

type replayStreamClient struct {
	path      string
	token     psclient.AccessToken
	calls     atomic.Int32
	lineCount atomic.Int32
}

func (c *replayStreamClient) SetToken(token psclient.AccessToken) {
	c.token = token
}

func (c *replayStreamClient) SyncStream(ctx context.Context, _ *psclient.SyncStreamRequest, handler psclient.LineHandler) error {
	c.calls.Add(1)
	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		copyLine := make([]byte, len(line))
		copy(copyLine, line)
		c.lineCount.Add(1)
		if err := handler(copyLine); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

func (c *replayStreamClient) GetWriteCheckpoint(context.Context, psclient.ClientID) (psclient.WriteCheckpoint, error) {
	return "0", nil
}

type syncerTokenSource struct {
	token syncer.AccessToken
}

func (s *syncerTokenSource) GetAccessToken(context.Context) (syncer.AccessToken, error) {
	return s.token, nil
}

func (s *syncerTokenSource) ForceRefreshAccessToken(context.Context) (syncer.AccessToken, error) {
	return s.token, nil
}

type uploaderTokenSource struct {
	token psclient.AccessToken
}

func (s *uploaderTokenSource) GetAccessToken(context.Context) (psclient.AccessToken, error) {
	return s.token, nil
}

type countingMutationHandler struct {
	calls atomic.Int32
}

func (h *countingMutationHandler) Handle(context.Context, psdb.Mutation) error {
	h.calls.Add(1)
	return nil
}

type flakyMutationHandler struct {
	calls      atomic.Int32
	failFirstN int32
}

func (h *flakyMutationHandler) Handle(context.Context, psdb.Mutation) error {
	if h.calls.Add(1) <= h.failFirstN {
		return context.DeadlineExceeded
	}
	return nil
}

func waitUploaderEvent(t *testing.T, events <-chan uploader.Event, timeout time.Duration) uploader.Event {
	t.Helper()
	select {
	case ev, ok := <-events:
		if !ok {
			t.Fatal("events channel closed before event")
		}
		return ev
	case <-time.After(timeout):
		t.Fatal("timed out waiting for uploader event")
		return nil
	}
}

func waitUntil(t *testing.T, timeout time.Duration, pred func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pred() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func openPowerSyncTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	if err := extension.Register(); err != nil {
		t.Fatalf("extension.Register() error = %v", err)
	}
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "powersync-integration.sqlite")
	db, err := sqlite.OpenBare(ctx, path)
	if err != nil {
		t.Fatalf("sqlite.OpenBare() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := extension.ApplySchema(ctx, db); err != nil {
		t.Fatalf("extension.ApplySchema() error = %v", err)
	}
	return db
}

func seedLocalBucket(t *testing.T, db *sqlite.DB, last, target int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES (?, ?, ?)", string(psdb.LocalBucket), last, target); err != nil {
		t.Fatalf("seed local bucket: %v", err)
	}
}

func insertCrud(t *testing.T, db *sqlite.DB, id int64, txID *int64, data string) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.Exec(ctx, "INSERT INTO ps_crud (id, tx_id, data) VALUES (?, ?, ?)", id, txID, data); err != nil {
		t.Fatalf("insert crud row: %v", err)
	}
}
