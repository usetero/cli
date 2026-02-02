package upload_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/api/apitest"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/powersynctest"
	"github.com/usetero/cli/internal/upload"
)

func TestUploader_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns on context cancellation", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		uploader := upload.New(
			db,
			powersynctest.NewMockClient(),
			powersynctest.NewMockTokenRefresher("token"),
			&apitest.MockConversations{},
			&chattest.MockMessages{},
			logtest.New(t),
		)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := uploader.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v, want context.Canceled", err)
		}
	})

	t.Run("closes event channel on exit", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)

		uploader := upload.New(
			db,
			powersynctest.NewMockClient(),
			powersynctest.NewMockTokenRefresher("token"),
			&apitest.MockConversations{},
			&chattest.MockMessages{},
			logtest.New(t),
		)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		cancel()
		<-done

		// Drain any pending events, then verify channel is closed
		for {
			_, ok := <-uploader.Events()
			if !ok {
				break
			}
		}
	})

	t.Run("processes entry and completes batch", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		convID := uuid.New().String()
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"conversations","id":"`+convID+`","data":{"workspace_id":"ws-1","title":"Test"}}`)

		conversations := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, id uuid.UUID, workspaceID, title string) (*api.Conversation, error) {
				return &api.Conversation{ID: id.String()}, nil
			},
		}

		uploader := upload.New(
			db,
			powersynctest.NewMockClient(),
			powersynctest.NewMockTokenRefresher("token"),
			conversations,
			&chattest.MockMessages{},
			logtest.New(t),
		)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		select {
		case event := <-uploader.Events():
			syncEvent, ok := event.(upload.SyncingEvent)
			if !ok {
				t.Errorf("expected SyncingEvent, got %T", event)
			} else if syncEvent.ProcessedCount != 1 {
				t.Errorf("expected ProcessedCount=1, got %d", syncEvent.ProcessedCount)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for syncing event")
		}

		cancel()
		<-done

		queue := powersync.NewCrudQueue(db)
		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}
		if entry != nil {
			t.Error("entry should be deleted after successful upload")
		}
	})

	t.Run("skips unknown tables and completes batch", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"unknown_table","id":"row-1","data":{}}`)

		uploader := upload.New(
			db,
			powersynctest.NewMockClient(),
			powersynctest.NewMockTokenRefresher("token"),
			&apitest.MockConversations{},
			&chattest.MockMessages{},
			logtest.New(t),
		)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		select {
		case event := <-uploader.Events():
			if _, ok := event.(upload.SyncingEvent); !ok {
				t.Errorf("expected SyncingEvent, got %T", event)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for syncing event")
		}

		cancel()
		<-done

		queue := powersync.NewCrudQueue(db)
		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}
		if entry != nil {
			t.Error("unknown table entry should be deleted after batch completes")
		}
	})

	t.Run("emits stalled event on failure and recovered on success", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		ctx := context.Background()

		_, err := db.Exec(ctx, "INSERT INTO ps_buckets (name, last_op, target_op) VALUES ('$local', 0, 0)")
		if err != nil {
			t.Fatalf("setup bucket: %v", err)
		}

		convID := uuid.New().String()
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"conversations","id":"`+convID+`","data":{"workspace_id":"ws-1","title":"Test"}}`)

		callCount := 0
		conversations := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, id uuid.UUID, workspaceID, title string) (*api.Conversation, error) {
				callCount++
				if callCount <= 4 {
					return nil, errors.New("temporary error")
				}
				return &api.Conversation{ID: id.String()}, nil
			},
		}

		uploader := upload.New(
			db,
			powersynctest.NewMockClient(),
			powersynctest.NewMockTokenRefresher("token"),
			conversations,
			&chattest.MockMessages{},
			logtest.New(t),
		)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		// Stalled event after retries exhausted
		select {
		case event := <-uploader.Events():
			stalledEvent, ok := event.(upload.StalledEvent)
			if !ok {
				t.Errorf("expected StalledEvent, got %T", event)
			} else {
				if stalledEvent.Error == nil {
					t.Error("expected error in stalled event")
				}
				if stalledEvent.Table != "conversations" {
					t.Errorf("expected table=conversations, got %s", stalledEvent.Table)
				}
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for stalled event")
		}

		// Recovered event after success
		select {
		case event := <-uploader.Events():
			if _, ok := event.(upload.RecoveredEvent); !ok {
				t.Errorf("expected RecoveredEvent, got %T", event)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for recovered event")
		}

		// Syncing event
		select {
		case event := <-uploader.Events():
			if _, ok := event.(upload.SyncingEvent); !ok {
				t.Errorf("expected SyncingEvent, got %T", event)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for syncing event")
		}

		cancel()
		<-done
	})
}
