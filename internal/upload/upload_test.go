package upload_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
		logger := logtest.New(t)

		uploader := upload.New(db, &apitest.MockConversations{}, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := uploader.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v, want context.Canceled", err)
		}
	})

	t.Run("closes event channel on exit", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)

		uploader := upload.New(db, &apitest.MockConversations{}, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		cancel()
		<-done

		// Channel should be closed
		_, ok := <-uploader.Events()
		if ok {
			t.Error("event channel should be closed after Run exits")
		}
	})

	t.Run("processes entry and deletes after success", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)

		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"conversations","id":"conv-1","data":{"workspace_id":"ws-1","title":"Test"}}`)

		conversations := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, workspaceID, title string) (*api.Conversation, error) {
				return &api.Conversation{ID: "conv-1"}, nil
			},
		}

		uploader := upload.New(db, conversations, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		// Wait for syncing event
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

	t.Run("skips unknown tables and deletes entry", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)

		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"unknown_table","id":"row-1","data":{}}`)

		uploader := upload.New(db, &apitest.MockConversations{}, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		time.Sleep(200 * time.Millisecond)
		cancel()
		<-done

		queue := powersync.NewCrudQueue(db)
		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}
		if entry != nil {
			t.Error("unknown table entry should be deleted")
		}
	})

	t.Run("emits stalled event on failure and recovered on success", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDB(t)
		logger := logtest.New(t)

		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"conversations","id":"conv-1","data":{"workspace_id":"ws-1","title":"Test"}}`)

		callCount := 0
		conversations := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, workspaceID, title string) (*api.Conversation, error) {
				callCount++
				// Fail first 4 attempts (initial + 3 retries), succeed on 5th
				if callCount <= 4 {
					return nil, errors.New("temporary error")
				}
				return &api.Conversation{ID: "conv-1"}, nil
			},
		}

		uploader := upload.New(db, conversations, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		// Should get stalled event after retries exhausted
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

		// Should get recovered event after success
		select {
		case event := <-uploader.Events():
			if _, ok := event.(upload.RecoveredEvent); !ok {
				t.Errorf("expected RecoveredEvent, got %T", event)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for recovered event")
		}

		// Should get syncing event
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
