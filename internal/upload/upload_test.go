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

		db := powersynctest.OpenTestDBWithSchema(t)
		logger := logtest.New(t)

		uploader := upload.New(db, &apitest.MockConversations{}, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := uploader.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v, want context.Canceled", err)
		}
	})

	t.Run("processes conversation entry and deletes after success", func(t *testing.T) {
		t.Parallel()

		db := powersynctest.OpenTestDBWithSchema(t)
		logger := logtest.New(t)

		// Insert a conversations entry
		powersynctest.InsertCrudEntry(t, db, 1, nil, `{"op":"PUT","type":"conversations","id":"conv-1","data":{"workspace_id":"ws-1","title":"Test"}}`)

		conversations := &apitest.MockConversations{
			CreateFunc: func(ctx context.Context, workspaceID, title string) (*api.Conversation, error) {
				return &api.Conversation{ID: "conv-1"}, nil
			},
		}

		uploader := upload.New(db, conversations, &chattest.MockMessages{}, logger)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Run in background
		done := make(chan error)
		go func() {
			done <- uploader.Run(ctx)
		}()

		// Wait for processing
		time.Sleep(200 * time.Millisecond)
		cancel()
		<-done

		// Verify entry was deleted
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

		db := powersynctest.OpenTestDBWithSchema(t)
		logger := logtest.New(t)

		// Insert an entry for unknown table
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

		// Entry should be deleted to avoid blocking queue
		queue := powersync.NewCrudQueue(db)
		entry, err := queue.GetNextEntry(context.Background())
		if err != nil {
			t.Fatalf("GetNextEntry() error = %v", err)
		}
		if entry != nil {
			t.Error("unknown table entry should be deleted")
		}
	})
}
